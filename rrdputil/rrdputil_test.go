package rrdputil

import (
	"fmt"
	"testing"
)

func TestGetRrdpNotificationAndRrdpSnapshot(t *testing.T) {
	notificationModel, err := GetRrdpNotification("https://rrdp.apnic.net/notification.xml")
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println("get notification ok")

	err = CheckRrdpNotification(&notificationModel)
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println("will get snapshot:", notificationModel.Snapshot.Uri)
	snapshotModel, err := GetRrdpSnapshot(notificationModel.Snapshot.Uri)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println("get snapshot ok")

	err = CheckRrdpSnapshot(&snapshotModel, &notificationModel)
	if err != nil {
		fmt.Println(err)
		return
	}

	for _, delta := range notificationModel.MapSerialDeltas {

		deltaModel, err := GetRrdpDelta(delta.Uri)
		if err != nil {
			fmt.Println(err)
			return
		}
		fmt.Println("get delta ok")

		err = CheckRrdpDelta(&deltaModel, &notificationModel)
		if err != nil {
			fmt.Println(err)
			return
		}
	}
	fmt.Println("all ok")
}
