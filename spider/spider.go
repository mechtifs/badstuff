package spider

import (
	"log"
	"sync"

	"badstuff/requests"
)

type Spider[T any] struct {
	Generate   func(chan string)
	Parse      func(*requests.Response) T
	Process    func(T)
	Finalize   func()
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
			c <- s.Parse(s.fetch(url))
			s.newTask()
		}()
	} else {
		s.resCh <- nil
	}
	s.mutex.Unlock()
}

func (s *Spider[T]) Run() {
	s.resCh = make(chan chan T, 2*s.NParallels)
	s.urlCh = make(chan string, 1)
	if s.MaxRetry == 0 {
		s.MaxRetry = 65535
	}
	if s.NParallels == 0 {
		s.NParallels = 256
	}
	s.mutex = &sync.Mutex{}

	go s.Generate(s.urlCh)

	for range s.NParallels {
		go s.newTask()
	}

	for c := range s.resCh {
		if c == nil {
			break
		}
		s.Process(<-c)
	}

	if s.Finalize != nil {
		s.Finalize()
	}
}
