package jobs

type Job struct {
	ID        int
	Processor IProcessor
}

type Worker struct {
	ID             int
	jobs           chan *Job
	dispatchStatus chan *DispatchStatus
	Quit           chan bool
}

func CreateNewWorker(id int, workerQueue chan *Worker, jobQueue chan *Job, dStatus chan *DispatchStatus) *Worker {
	w := &Worker{
		ID:             id,
		jobs:           jobQueue,
		dispatchStatus: dStatus,
	}

	go func() { workerQueue <- w }()
	return w
}

func (w *Worker) Start() {
	go func() {
		for {
			select {
			case job := <-w.jobs:
				job.Processor.Run()
				w.dispatchStatus <- &DispatchStatus{Type: "worker", ID: w.ID, Status: "quit"}
				w.Quit <- true
			case <-w.Quit:
				return
			}
		}
	}()
}

type DispatchStatus struct {
	Type   string
	ID     int
	Status string
}

type Dispatcher struct {
	jobCounter     int                  // internal counter for number of jobs
	jobQueue       chan *Job            // channel of jobs submitted by main()
	dispatchStatus chan *DispatchStatus // channel for job/worker status reports
	workQueue      chan *Job            // channel of work dispatched
	workerQueue    chan *Worker         // channel of workers
}

func CreateNewDispatcher() *Dispatcher {
	d := &Dispatcher{
		jobCounter:     0,
		jobQueue:       make(chan *Job),
		dispatchStatus: make(chan *DispatchStatus),
		workQueue:      make(chan *Job),
		workerQueue:    make(chan *Worker),
	}
	return d
}

func (d *Dispatcher) Start(numWorkers int) {
	// Create numWorkers:
	for i := 0; i < numWorkers; i++ {
		worker := CreateNewWorker(i, d.workerQueue, d.workQueue, d.dispatchStatus)
		worker.Start()
	}

	// wait for work to be added then pass it off.
	go func() {
		for {
			select {
			case job := <-d.jobQueue:
				d.workQueue <- job
			case ds := <-d.dispatchStatus:
				if ds.Type == "worker" {
					if ds.Status == "quit" {
						d.jobCounter--
					}
				}
			}
		}
	}()
}

func (d *Dispatcher) AddJob(je IProcessor) {
	j := &Job{ID: d.jobCounter, Processor: je}
	go func() { d.jobQueue <- j }()
	d.jobCounter++
}

func (d *Dispatcher) Finished() bool {
	if d.jobCounter < 1 {
		return true
	} else {
		return false
	}
}
