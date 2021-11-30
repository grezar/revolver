package reporting

import (
	"os"
	"runtime"
	"sync"

	"github.com/olekukonko/tablewriter"
)

const (
	Success = "SUCCESS"
	Skip    = "SKIP"
	Error   = "ERROR"
)

func Run(f func(r *R)) {
	ctx := newReportContext()
	r := &R{
		barrier: make(chan bool),
		done:    make(chan bool),
		context: ctx,
	}
	go rRunner(r, f)
	<-r.done
	r.Render()
}

type R struct {
	mu         sync.RWMutex
	name       string
	status     string
	summary    string
	err        string
	parent     *R
	children   []*R
	sub        []*R
	isParallel bool
	context    *reportContext
	barrier    chan bool
	done       chan bool
}

func (r *R) Run(name string, f func(r *R)) {
	ctx := newReportContext()
	child := &R{
		barrier: make(chan bool),
		done:    make(chan bool),
		name:    name,
		parent:  r,
		context: ctx,
	}
	go rRunner(child, f)
	<-child.done
	r.appendChild(child)
}

func rRunner(r *R, fn func(r *R)) {
	defer func() {
		if len(r.sub) > 0 {
			// Run parallel sub reports.
			// Decrease the running count for this report.
			r.context.release()
			// Release the parallel sub reports.
			close(r.barrier)
			// Wait for sub reports to complete.
			for _, sub := range r.sub {
				<-sub.done
			}
			if !r.isParallel {
				// Reacuire the count for sequential reports. See comment in Run.
				r.context.waitParallel()
			}
		} else if r.isParallel {
			// Only release the count for this report if it was run as a parallel
			// report. See comment in Run method.
			r.context.release()
		}

		r.done <- true
	}()

	fn(r)
}

func (r *R) Parallel() {
	if r.isParallel {
		panic("reporter: r.Parallel called multiple times")
	}
	r.isParallel = true
	r.parent.sub = append(r.parent.sub, r)
	r.done <- true     // Release calling report
	<-r.parent.barrier // Wait for the parent report to complete
	r.context.waitParallel()
}

func (r *R) appendChild(child *R) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.children = append(r.children, child)
}

type reportContext struct {
	mu            sync.Mutex
	running       int
	startParallel chan bool
	numWaiting    int
	maxParallel   int
}

func newReportContext() *reportContext {
	return &reportContext{
		startParallel: make(chan bool),
		maxParallel:   runtime.GOMAXPROCS(0),
		running:       1,
	}
}

func (c *reportContext) waitParallel() {
	c.mu.Lock()
	if c.running < c.maxParallel {
		c.running++
		c.mu.Unlock()
		return
	}
	c.numWaiting++
	c.mu.Unlock()
	<-c.startParallel
}

func (c *reportContext) release() {
	c.mu.Lock()
	if c.numWaiting == 0 {
		c.running--
		c.mu.Unlock()
		return
	}
	c.numWaiting--
	c.mu.Unlock()
	c.startParallel <- true
}

func (r *R) Render() {
	var rows [][]string
	for _, rotation := range r.children {
		rows = append(rows, []string{rotation.name, "", "", "", ""})
		for _, provider := range rotation.children {
			rows = append(rows, []string{"", provider.name, provider.status, provider.summary, provider.err})
		}
	}

	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"ROTATION", "PROVIDER", "STATUS", "SUMMARY", "ERROR"})
	table.SetAutoFormatHeaders(true)
	table.SetHeaderAlignment(tablewriter.ALIGN_LEFT)
	table.SetAlignment(tablewriter.ALIGN_LEFT)
	table.SetCenterSeparator("")
	table.SetColumnSeparator("")
	table.SetRowSeparator("")
	table.SetHeaderLine(false)
	table.SetBorder(false)
	table.SetTablePadding("\t")
	table.SetNoWhiteSpace(true)

	var bgColor int
	for _, row := range rows {
		switch row[2] {
		case "UPDATED":
			bgColor = tablewriter.BgMagentaColor
		case "SKIP":
			bgColor = tablewriter.BgCyanColor
		case "ERROR":
			bgColor = tablewriter.BgRedColor
		case "SUCCESS":
			bgColor = tablewriter.BgGreenColor
		}
		table.Rich(row, []tablewriter.Colors{{}, {}, {tablewriter.FgHiBlackColor, tablewriter.Bold, bgColor}, {}})
	}

	table.Render()
}

func (r *R) Summary(summary string) {
	r.summary = summary
}

func (r *R) Success() {
	r.status = Success
}

func (r *R) Skip() {
	r.status = Skip
}

func (r *R) Fail(err error) {
	r.err = err.Error()
	r.status = Error
}
