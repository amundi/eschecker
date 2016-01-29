package worker

var G_WorkQueue = make(chan WorkRequest)

//dispatcher
var G_WorkerQueue chan chan WorkRequest
var G_WorkerSlice []Worker

//interface used to update and display the stats
type WorkRequest interface {
	DoRequest()
}

type Worker struct {
	ID          int
	Work        chan WorkRequest
	WorkerQueue chan chan WorkRequest
	QuitChan    chan bool
}

func StartDispatcher(nbWorkers int) {
	G_WorkerQueue = make(chan chan WorkRequest, nbWorkers)
	G_WorkerSlice = make([]Worker, nbWorkers)

	//create workers
	for i := 0; i < nbWorkers; i++ {
		G_WorkerSlice[i] = newWorker(i+1, G_WorkerQueue)
		G_WorkerSlice[i].Start()
	}

	go func() {
		for {
			select {
			//a request is sent to the WorkQueue
			case work := <-G_WorkQueue:
				go func() {
					//a worker is pulled from the workerqueue
					worker := <-G_WorkerQueue
					//the worker handles the work
					worker <- work
				}()
			}
		}
	}()
}

func StopAllWorkers(nbWorkers uint32) {
	for i := uint32(0); i < nbWorkers; i++ {
		G_WorkerSlice[i].Stop()
	}
}

func newWorker(id int, workerQueue chan chan WorkRequest) Worker {
	worker := Worker{
		ID:          id,
		Work:        make(chan WorkRequest),
		WorkerQueue: workerQueue,
		QuitChan:    make(chan bool),
	}
	return worker
}

func (w Worker) Start() {
	go func() {
		for {
			// Add ourselves into the worker queue.
			w.WorkerQueue <- w.Work

			select {
			case work := <-w.Work:
				//do the request, whatever it is
				work.DoRequest()
			case <-w.QuitChan:
				// We have been asked to stop.
				return
			}
		}
	}()
}

func (w Worker) Stop() {
	go func() {
		w.QuitChan <- true
	}()
}
