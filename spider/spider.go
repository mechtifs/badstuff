package spider

import (
	"log"
	"sync"

	"badstuff/requests"
)

type Workflow[T any] interface {
	Init()
	Generate(chan string)
	Parse(*requests.Response) T
	Process(T)
	Finalize()
}

type Spider[T any] struct {
	Workflow   Workflow[T]
	Session    *requests.Session
	NParallels int
	MaxRetry   int

	resCh chan chan T
	urlCh chan string
	mutex *sync.Mutex
}

func (s *Spider[T]) fetch(url string) *requests.Response {
	log.Println("Fetching", url)
	for i := range s.MaxRetry + 1 {
		if i != 0 {
			log.Println("Retrying", url, "*", i)
		}
		r, err := s.Session.Get(url, nil)
		if err != nil {
			log.Println(err.Error())
			continue
		}
		if r.StatusCode < 500 {
			log.Println("Finished", url)
			return r
		}
	}
	log.Println("Skipped", url)
	return nil
}

func (s *Spider[T]) newTask() {
	s.mutex.Lock()
	url, ok := <-s.urlCh
	if ok {
		c := make(chan T, 1)
		s.resCh <- c
		go func() {
			c <- s.Workflow.Parse(s.fetch(url))
			s.newTask()
		}()
	} else {
		s.resCh <- nil
	}
	s.mutex.Unlock()
}

func (s *Spider[T]) Run() {
	s.Workflow.Init()

	s.resCh = make(chan chan T, 2*s.NParallels)
	s.urlCh = make(chan string, 1)
	if s.MaxRetry == 0 {
		s.MaxRetry = 65535
	}
	if s.NParallels == 0 {
		s.NParallels = 32
	}
	s.mutex = &sync.Mutex{}

	go s.Workflow.Generate(s.urlCh)

	for range s.NParallels {
		go s.newTask()
	}

	for c := range s.resCh {
		if c == nil {
			break
		}
		s.Workflow.Process(<-c)
	}

	s.Workflow.Finalize()
}
