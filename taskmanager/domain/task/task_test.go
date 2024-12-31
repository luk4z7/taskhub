package task_test

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/luk4z7/taskmanager/domain/task"
	mocktask "github.com/luk4z7/taskmanager/domain/task/mock"
	"github.com/luk4z7/taskmanager/domain/user"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestTaskValidationOnSave(t *testing.T) {
	ctrl := gomock.NewController(t)

	testCase := []struct {
		id      int
		request task.Task
		want    error
	}{
		{1, task.MarshalTask("123", time.Now(), user.User("user@gmail.com")), nil},
		{2, task.MarshalTask(gofakeit.HipsterParagraph(3, 5, 100, "\n"), time.Now(), user.User("user@gmail.com")), errors.New("summary exceded the limit of 2500")},
		{3, task.MarshalTask(gofakeit.HipsterParagraph(3, 5, 10, "\n"), time.Now(), user.User("user@gmail.com")), nil},
	}

	for _, tc := range testCase {
		t.Run(fmt.Sprintf("ID : %d", tc.id), func(t *testing.T) {
			taskMock := mocktask.NewMockTaskRepository(ctrl)
			ctx := context.Background()

			if tc.want == nil {
				taskMock.EXPECT().AddTask(ctx, gomock.Any()).Return(tc.want)
			}

			tsk := task.New(taskMock)
			assert.Equal(t, tsk.Save(ctx, tc.request, tc.request.CreatedBy()), tc.want)
		})
	}
}

func TestTaskList(t *testing.T) {
	ctrl := gomock.NewController(t)

	testCase := []struct {
		id      int
		request user.Role
		result  []task.Task
		want    error
	}{
		{1, user.Role("manager"), []task.Task{
			0: task.MarshalTask(gofakeit.HipsterParagraph(3, 5, 10, "\n"), time.Now(), user.User("user@gmail.com")),
			1: task.MarshalTask(gofakeit.HipsterParagraph(3, 5, 10, "\n"), time.Now(), user.User("admin@gmail.com")),
		}, nil},
	}

	for _, tc := range testCase {
		t.Run(fmt.Sprintf("ID : %d", tc.id), func(t *testing.T) {
			taskMock := mocktask.NewMockTaskRepository(ctrl)
			ctx := context.Background()

			taskMock.EXPECT().List(ctx, tc.request).Return(tc.result, tc.want)

			tsk := task.New(taskMock)
			res, err := tsk.List(ctx, tc.request)
			assert.Equal(t, err, tc.want)
			assert.Equal(t, res, tc.result)
		})
	}

}
