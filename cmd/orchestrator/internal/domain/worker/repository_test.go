package worker

import (
	"go.uber.org/mock/gomock"
	"testing"
)

const ip = "127.0.0.1"

func TestServiceAddWorker(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	store := NewMockStore(ctrl)
	store.EXPECT().Add(Worker{IP: ip}).MinTimes(1)

	srv := NewService(store)
	_ = srv.AddWorker(ip)
}

func TestServiceRemoveWorker(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	store := NewMockStore(ctrl)
	store.EXPECT().Remove(Worker{IP: ip}).MinTimes(1)

	srv := NewService(store)
	_ = srv.RemoveWorker(ip)
}

func TestServiceListWorker(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	store := NewMockStore(ctrl)
	store.EXPECT().List().MinTimes(1)

	srv := NewService(store)
	_ = srv.List()
}
