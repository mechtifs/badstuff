package spider

import (
	"log"
	"sync"

	"badstuff/requests"
)

type Spider struct {
	Generate       func(chan string)
	Parse          func(*requests.Response) interface{}
	Process        func(interface{})
	NParallels     int
	MaxRetry       int
	SessionOptions *requests.SessionOptions

	resCh   chan chan interface{}
	urlCh   chan string
	mutex   *sync.Mutex
	session *requests.Session
}

func (s *Spider) fetch(url string) *requests.Response {
	log.Println("Fetching", url)
	for i := 0; i <= s.MaxRetry; i++ {
		if i != 0 {
			log.Println("Retrying", url, "*", i)
		}
		r, err := s.session.Get(url, nil)
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

func (s *Spider) newTask() {
	s.mutex.Lock()
	url, ok := <-s.urlCh
	if ok {
		c := make(chan interface{}, 1)
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

func (s *Spider) Run() {
	s.resCh = make(chan chan interface{}, 2*s.NParallels)
	s.urlCh = make(chan string, 1)
	if s.MaxRetry == 0 {
		s.MaxRetry = 65535
	}
	if s.NParallels == 0 {
		s.NParallels = 256
	}
	s.session = requests.NewSession(s.SessionOptions)
	s.mutex = &sync.Mutex{}

	go s.Generate(s.urlCh)

	for i := 0; i < s.NParallels; i++ {
		go s.newTask()
	}

	for c := range s.resCh {
		if c == nil {
			break
		}
		s.Process(<-c)
	}
}
