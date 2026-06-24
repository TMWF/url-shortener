package pool_test

import (
	"errors"
	"testing"

	"github.com/TMWF/url-shortener/internal/pool"
)

// DummyStruct удовлетворяет интерфейсу Resetter для тестирования
type DummyStruct struct {
	Value int
}

func (d *DummyStruct) Reset() {
	d.Value = 0
}

// TestNew_NilFactory проверяет, что передача nil-фабрики возвращает ошибку
func TestNew_NilFactory(t *testing.T) {
	p, err := pool.New[*DummyStruct](nil)

	if !errors.Is(err, pool.ErrNilFactory) {
		t.Fatalf("ожидалась ошибка %v, получена %v", pool.ErrNilFactory, err)
	}

	if p != nil {
		t.Error("пул должен быть nil при ошибке инициализации")
	}
}

// TestPool_Lifecycle проверяет стандартный цикл работы с пулом: Get, Put, Reset
func TestPool_Lifecycle(t *testing.T) {
	p, err := pool.New[*DummyStruct](func() *DummyStruct {
		return &DummyStruct{Value: 42}
	})
	if err != nil {
		t.Fatalf("не удалось инициализировать пул: %v", err)
	}

	obj := p.Get()
	if obj.Value != 42 {
		t.Errorf("ожидалось начальное значение 42, получено %d", obj.Value)
	}

	obj.Value = 100
	p.Put(obj)

	obj2 := p.Get()
	if obj2.Value != 0 {
		t.Errorf("ожидалось, что объект будет сброшен (0), получено %d", obj2.Value)
	}
}
