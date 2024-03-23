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
	cPool *sync.Pool
	mutex *sync.Mutex
	wg    *sync.WaitGroup
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

func (s *Spider[T]) worker() {
	defer s.wg.Done()
	for {
		s.mutex.Lock()
		if url, ok := <-s.urlCh; ok {
			c := s.cPool.Get().(chan T)
			s.resCh <- c
			s.mutex.Unlock()
			s.wg.Add(1)
			go func() {
				c <- s.Workflow.Parse(s.fetch(url))
				s.wg.Done()
			}()
		} else {
			s.mutex.Unlock()
			break
		}
	}
}

func (s *Spider[T]) Run() {
	s.Workflow.Init()

	s.resCh = make(chan chan T, 2*s.NParallels)
	s.urlCh = make(chan string, 2*s.NParallels)
	if s.MaxRetry <= 0 {
		s.MaxRetry = 65535
	}
	if s.NParallels <= 0 {
		s.NParallels = 32
	}
	s.cPool = &sync.Pool{
		New: func() any {
			return make(chan T, 1)
		},
	}
	s.mutex = &sync.Mutex{}
	s.wg = &sync.WaitGroup{}

	go s.Workflow.Generate(s.urlCh)

	go func() {
		for c := range s.resCh {
			val := <-c
			s.cPool.Put(c)
			s.Workflow.Process(val)
		}
		s.wg.Done()
	}()

	s.wg.Add(s.NParallels)
	for range s.NParallels {
		go s.worker()
	}
	s.wg.Wait()

	s.wg.Add(1)
	close(s.resCh)
	s.wg.Wait()

	s.Workflow.Finalize()
}
