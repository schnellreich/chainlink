package web

import (
	"errors"

	"github.com/gin-gonic/gin"
	"github.com/manyminds/api2go/jsonapi"
	"github.com/smartcontractkit/chainlink/services"
	"github.com/smartcontractkit/chainlink/store/models"
	"github.com/smartcontractkit/chainlink/store/orm"
	"github.com/smartcontractkit/chainlink/store/presenters"
)

// JobSpecsController manages JobSpec requests.
type JobSpecsController struct {
	App services.Application
}

// Index lists JobSpecs, one page at a time.
// Example:
//  "<application>/specs?size=1&page=2"
func (jsc *JobSpecsController) Index(c *gin.Context, size, page, offset int) {
	var order orm.SortType
	if c.Query("sort") == "-createdAt" {
		order = orm.Descending
	} else {
		order = orm.Ascending
	}

	jobs, count, err := jsc.App.GetStore().JobsSorted(order, offset, size)
	pjs := make([]presenters.JobSpec, len(jobs))
	for i, j := range jobs {
		pjs[i] = presenters.JobSpec{JobSpec: j}
	}

	paginatedResponse(c, "Jobs", size, page, pjs, count, err)
}

// Create adds validates, saves, and starts a new JobSpec.
// Example:
//  "<application>/specs"
func (jsc *JobSpecsController) Create(c *gin.Context) {
	js := models.NewJob()
	if err := c.ShouldBindJSON(&js); err != nil {
		publicError(c, 400, err)
	} else if err := services.ValidateJob(js, jsc.App.GetStore()); err != nil {
		publicError(c, 400, err)
	} else if err = jsc.App.AddJob(js); err != nil {
		c.AbortWithError(500, err)
	} else if doc, err := jsonapi.Marshal(presenters.JobSpec{JobSpec: js, Runs: []presenters.JobRun{}}); err != nil {
		c.AbortWithError(500, err)
	} else {
		c.Data(200, MediaType, doc)
	}
}

// Show returns the details of a JobSpec.
// Example:
//  "<application>/specs/:SpecID"
func (jsc *JobSpecsController) Show(c *gin.Context) {
	id := c.Param("SpecID")
	if j, err := jsc.App.GetStore().FindJob(id); err == orm.ErrorNotFound {
		publicError(c, 404, errors.New("JobSpec not found"))
	} else if err != nil {
		c.AbortWithError(500, err)
	} else if runs, err := jsc.App.GetStore().JobRunsFor(j.ID); err != nil {
		c.AbortWithError(500, err)
	} else if doc, err := marshalSpecFromJSONAPI(j, runs); err != nil {
		c.AbortWithError(500, err)
	} else {
		c.JSON(200, doc)
	}
}

func marshalSpecFromJSONAPI(j models.JobSpec, runs []models.JobRun) (*jsonapi.Document, error) {
	pruns := make([]presenters.JobRun, len(runs))
	for i, r := range runs {
		pruns[i] = presenters.JobRun{r}
	}
	p := presenters.JobSpec{JobSpec: j, Runs: pruns}
	doc, err := jsonapi.MarshalToStruct(p, nil)
	return doc, err
}
