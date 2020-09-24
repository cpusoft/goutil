package rrdputil

import (
	"fmt"
	"testing"
)

func TestGetRrdpNotification(t *testing.T) {
	notificationModel, err := GetRrdpNotification("https://rpki.idnic.net/rrdp/notify.xml")
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
}
func TestGetRrdpSnapshot(t *testing.T) {
	url := `https://rrdp.arin.net/8fe05c2e-047d-49e7-8398-cd4250a572b1/7677/snapshot.xml`
	fmt.Println("will get snapshot:", url)
	snapshotModel, err := GetRrdpSnapshot(url)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println("get snapshot ok:", snapshotModel)
	/*
		err = CheckRrdpSnapshot(&snapshotModel, &notificationModel)
		if err != nil {
			fmt.Println(err)
			return
		}

			err = SaveRrdpSnapshotToFiles(&snapshotModel, `G:\Download\rrdp`)
			if err != nil {
				fmt.Println(err)
				return
			}

			for i, _ := range notificationModel.Deltas {

				deltaModel, err := GetRrdpDelta(notificationModel.Deltas[i].Uri)
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
				err = SaveRrdpDeltaToFiles(&deltaModel, `G:\Download\rrdp`)
				if err != nil {
					fmt.Println(err)
					return
				}
			}
	*/
	fmt.Println("all ok")
}
