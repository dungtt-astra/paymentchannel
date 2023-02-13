//
// machine_mock_test.go
//
// Distributed under terms of the MIT license.
//

package mock_machine_test

import (
	context "context"
	machine "github.com/dungtt-astra/paymentchannel/machine"
	"log"
	"testing"
	"time"

	mock_machine "github.com/dungtt-astra/paymentchannel/mock_machine"
	gomock "github.com/golang/mock/gomock"
)

func testExecute(t *testing.T, client machine.MachineClient) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	instructions := []*machine.Instruction{}
	instructions = append(instructions, &machine.Instruction{Operand: 5, Operator: "PUSH"})
	instructions = append(instructions, &machine.Instruction{Operand: 6, Operator: "PUSH"})
	instructions = append(instructions, &machine.Instruction{Operator: "MUL"})

	stream, err := client.Execute(ctx)
	if err != nil {
		log.Fatalf("%v.Execute(%v) = _, %v: ", client, ctx, err)
	}
	for _, instruction := range instructions {
		if err := stream.Send(instruction); err != nil {
			log.Fatalf("%v.Send(%v) = %v: ", stream, instruction, err)
		}
	}
	result, err := stream.Recv()
	if err != nil {
		log.Fatalf("%v.Recv() got error %v, want %v", stream, err, nil)
	}

	got := result.GetOutput()
	want := float32(30)
	if got != want {
		t.Errorf("got %v, want %v", got, want)
	}
}

func TestExecute(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockMachineClient := mock_machine.NewMockMachineClient(ctrl)

	mockClientStream := mock_machine.NewMockMachine_ExecuteClient(ctrl)
	mockClientStream.EXPECT().Send(gomock.Any()).Return(nil).AnyTimes()
	mockClientStream.EXPECT().Recv().Return(&machine.Result{Output: 30}, nil)

	mockMachineClient.EXPECT().Execute(
		gomock.Any(), // context
	).Return(mockClientStream, nil)

	testExecute(t, mockMachineClient)
}
