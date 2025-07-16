package scraper

import (
	"runtime"
)

type Storage interface {
	Clean()
	Store(metricName string, value float64)
	Metrics() map[string]float64
}

type Scraper struct {
	count   int64
	storage Storage
}

func New(storage Storage) *Scraper {
	return &Scraper{
		count:   0,
		storage: storage,
	}
}

func (s *Scraper) Metrics() map[string]float64 {
	return s.storage.Metrics()
}

func (s *Scraper) Count() int64 {
	return s.count
}

func (s *Scraper) Reset() {
	s.storage.Clean()
	s.count = 0
}

func (s *Scraper) Scrap() {
	memStats := runtime.MemStats{}
	runtime.ReadMemStats(&memStats)

	s.storage.Store("Alloc", float64(memStats.Alloc))
	s.storage.Store("BuckHashSys", float64(memStats.BuckHashSys))
	s.storage.Store("Frees", float64(memStats.Frees))
	s.storage.Store("GCCPUFraction", memStats.GCCPUFraction)
	s.storage.Store("GCSys", float64(memStats.GCSys))
	s.storage.Store("HeapAlloc", float64(memStats.HeapAlloc))
	s.storage.Store("HeapIdle", float64(memStats.HeapIdle))
	s.storage.Store("HeapInuse", float64(memStats.HeapInuse))
	s.storage.Store("HeapObjects", float64(memStats.HeapObjects))
	s.storage.Store("HeapReleased", float64(memStats.HeapReleased))
	s.storage.Store("HeapSys", float64(memStats.HeapSys))
	s.storage.Store("LastGC", float64(memStats.LastGC))
	s.storage.Store("Lookups", float64(memStats.Lookups))
	s.storage.Store("MCacheInuse", float64(memStats.MCacheInuse))
	s.storage.Store("MCacheSys", float64(memStats.MCacheSys))
	s.storage.Store("MSpanInuse", float64(memStats.MSpanInuse))
	s.storage.Store("MSpanSys", float64(memStats.MSpanSys))
	s.storage.Store("Mallocs", float64(memStats.Mallocs))
	s.storage.Store("NextGC", float64(memStats.NextGC))
	s.storage.Store("NumForcedGC", float64(memStats.NumForcedGC))
	s.storage.Store("NumGC", float64(memStats.NumGC))
	s.storage.Store("OtherSys", float64(memStats.OtherSys))
	s.storage.Store("PauseTotalNs", float64(memStats.PauseTotalNs))
	s.storage.Store("StackInuse", float64(memStats.StackInuse))
	s.storage.Store("StackSys", float64(memStats.StackSys))
	s.storage.Store("Sys", float64(memStats.Sys))
	s.storage.Store("TotalAlloc", float64(memStats.TotalAlloc))

	s.count++
}
