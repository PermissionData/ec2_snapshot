package main

import (
	"errors"
	"fmt"
	"math/rand"
	"testing"
	"time"

	"github.com/PermissionData/ec2_snapshot/mock_ec2iface"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/golang/mock/gomock"
)

func getMocks(t *testing.T) (*mock_ec2iface.MockEC2API, *gomock.Controller) {
	ctrl := gomock.NewController(t)
	return mock_ec2iface.NewMockEC2API(ctrl), ctrl
}

func TestRemoveOldIMage(t *testing.T) {
	mockEC2iface, ctrl := getMocks(t)
	defer ctrl.Finish()

	var getTimeSecondsBeforeNowAsString = func(s int64) *string {
		t := time.Now().Add(-time.Duration(s) * time.Second).Format(time.RFC3339)
		return &t
	}

	var filters = []*ec2.Filter{
		{
			Name: aws.String("owner-id"),
			Values: []*string{
				aws.String("533779774295"),
			},
		},
	}

	var happyPathTests = []struct {
		s            *svcEC2
		newImageID   string
		awsImages    []*ec2.Image
		deletes      []int
		AWSSnapshots []*ec2.Snapshot
	}{
		{
			s: &svcEC2{
				svc: mockEC2iface,
				imageNameWithoutTimestamp: "testing1.bak",
				imageName:                 "testing1.bak.1257894000",
				timeToSave:                604800,
				filter:                    filters,
			},
			newImageID: "ami-123456d",
			awsImages: []*ec2.Image{
				{
					ImageId:      aws.String("ami-123456a"),
					Name:         aws.String("testing1.bak.848590424"),
					CreationDate: getTimeSecondsBeforeNowAsString(604799),
				},
				{
					ImageId:      aws.String("ami-123456b"),
					Name:         aws.String("testing1.bak.438208309884"),
					CreationDate: getTimeSecondsBeforeNowAsString(604801),
				},
				{
					ImageId:      aws.String("ami-123456c"),
					Name:         aws.String("testing1.bak.4284932088"),
					CreationDate: getTimeSecondsBeforeNowAsString(604799),
				},
				{
					ImageId:      aws.String("ami-123456d"),
					Name:         aws.String("testing1.bak.993948322"),
					CreationDate: getTimeSecondsBeforeNowAsString(2),
				},
				{
					ImageId:      aws.String("ami-123456e"),
					Name:         aws.String("testing1.bak.3898349383"),
					CreationDate: getTimeSecondsBeforeNowAsString(1604802),
				},
			},
			deletes: []int{1, 4},
			AWSSnapshots: []*ec2.Snapshot{
				{
					SnapshotId: aws.String("test1"),
					Description: aws.String(
						"This snapshot is taken from ami-123456b",
					),
				},
				{
					SnapshotId: aws.String("test2"),
					Description: aws.String(
						"This snapshot is taken from ami-123456e",
					),
				},
			},
		},
		{
			s: &svcEC2{
				svc: mockEC2iface,
				imageNameWithoutTimestamp: "testing1.bak",
				imageName:                 "testing1.bak.1257894000",
				timeToSave:                604800,
				filter:                    filters,
			},
			newImageID: "ami-123456d",
			awsImages: []*ec2.Image{
				{
					ImageId:      aws.String("ami-123456a"),
					Name:         aws.String("testing1.bak.848590424"),
					CreationDate: getTimeSecondsBeforeNowAsString(604802),
				},
				{
					ImageId:      aws.String("ami-123456b"),
					Name:         aws.String("testing1.bak.438208309884"),
					CreationDate: getTimeSecondsBeforeNowAsString(604801),
				},
				{
					ImageId:      aws.String("ami-123456c"),
					Name:         aws.String("testing1.bak.4284932088"),
					CreationDate: getTimeSecondsBeforeNowAsString(900000),
				},
				{
					ImageId:      aws.String("ami-123456d"),
					Name:         aws.String("testing1.bak.993948322"),
					CreationDate: getTimeSecondsBeforeNowAsString(999999999),
				},
				{
					ImageId:      aws.String("ami-123456e"),
					Name:         aws.String("testing1.bak.3898349383"),
					CreationDate: getTimeSecondsBeforeNowAsString(1000000),
				},
			},
			deletes: []int{0, 1, 2, 3, 4},
			AWSSnapshots: []*ec2.Snapshot{
				{
					SnapshotId: aws.String("test1"),
					Description: aws.String(
						"This snapshot is taken from ami-123456b",
					),
				},
				{
					SnapshotId: aws.String("test2"),
					Description: aws.String(
						"This snapshot is taken from ami-123456e",
					),
				},
				{
					SnapshotId: aws.String("test3"),
					Description: aws.String(
						"This snapshot is taken from ami-123456a",
					),
				},
				{
					SnapshotId: aws.String("test4"),
					Description: aws.String(
						"This snapshot is taken from ami-123456c",
					),
				},
				{
					SnapshotId: aws.String("test5"),
					Description: aws.String(
						"This snapshot is taken from ami-123456d",
					),
				},
			},
		},
		{
			s: &svcEC2{
				svc: mockEC2iface,
				imageNameWithoutTimestamp: "testing1.bak",
				imageName:                 "testing1.bak.1257894000",
				timeToSave:                604800,
				filter:                    filters,
			},
			newImageID: "ami-123456d",
			awsImages: []*ec2.Image{
				{
					ImageId:      aws.String("ami-123456a"),
					Name:         aws.String("testing1.bak.848590424"),
					CreationDate: getTimeSecondsBeforeNowAsString(604799),
				},
				{
					ImageId:      aws.String("ami-123456b"),
					Name:         aws.String("testing2.bak.438208309884"),
					CreationDate: getTimeSecondsBeforeNowAsString(604801),
				},
				{
					ImageId:      aws.String("ami-123456c"),
					Name:         aws.String("testing1.bak.4284932088"),
					CreationDate: getTimeSecondsBeforeNowAsString(604799),
				},
				{
					ImageId:      aws.String("ami-123456d"),
					Name:         aws.String("testing1.bak.993948322"),
					CreationDate: getTimeSecondsBeforeNowAsString(2),
				},
				{
					ImageId:      aws.String("ami-123456e"),
					Name:         aws.String("testing3.bak.3898349383"),
					CreationDate: getTimeSecondsBeforeNowAsString(1604802),
				},
			},
			deletes: []int{},
			AWSSnapshots: []*ec2.Snapshot{
				{
					SnapshotId: aws.String("test1"),
					Description: aws.String(
						"This snapshot is taken from ami-123456b",
					),
				},
				{
					SnapshotId: aws.String("test2"),
					Description: aws.String(
						"This snapshot is taken from ami-123456e",
					),
				},
			},
		},
	}

	for _, test := range happyPathTests {
		mockEC2iface.EXPECT().DescribeImages(
			&ec2.DescribeImagesInput{Filters: filters},
		).Return(
			&ec2.DescribeImagesOutput{Images: test.awsImages},
			nil,
		)
		for index, i := range test.deletes {
			mockEC2iface.EXPECT().DeregisterImage(
				&ec2.DeregisterImageInput{
					ImageId: test.awsImages[i].ImageId,
					DryRun:  aws.Bool(false),
				},
			).Return(
				&ec2.DeregisterImageOutput{},
				nil,
			)
			mockEC2iface.EXPECT().DescribeSnapshots(
				&ec2.DescribeSnapshotsInput{
					Filters: filters,
				},
			).Return(
				&ec2.DescribeSnapshotsOutput{
					Snapshots: test.AWSSnapshots,
				},
				nil,
			)
			mockEC2iface.EXPECT().DeleteSnapshot(
				&ec2.DeleteSnapshotInput{
					SnapshotId: test.AWSSnapshots[index].SnapshotId,
					DryRun:     aws.Bool(false),
				},
			).Return(
				&ec2.DeleteSnapshotOutput{},
				nil,
			)
		}
		err := test.s.removeOldImage(test.newImageID)
		if err != nil {
			t.Errorf("Expect 'nil' got %v", err)
		}
	}

	var describeImagesErrorTest = []struct {
		s          *svcEC2
		newImageID string
	}{
		{
			s: &svcEC2{
				svc: mockEC2iface,
				imageNameWithoutTimestamp: "testing1.bak",
				imageName:                 "testing1.bak.1257894000",
				timeToSave:                604800,
				filter:                    filters,
			},
			newImageID: "ami-123456d",
		},
	}

	for _, test := range describeImagesErrorTest {
		mockEC2iface.EXPECT().DescribeImages(
			&ec2.DescribeImagesInput{Filters: filters},
		).Return(
			nil,
			errors.New("Some error blah blah"),
		)
		err := test.s.removeOldImage(test.newImageID)
		if err == nil {
			t.Error("Expect an error but got nil")
		}
	}

	var deregisterImageErrorTest = []struct {
		s            *svcEC2
		newImageID   string
		awsImages    []*ec2.Image
		deletes      []int
		AWSSnapshots []*ec2.Snapshot
	}{
		{
			s: &svcEC2{
				svc: mockEC2iface,
				imageNameWithoutTimestamp: "testing1.bak",
				imageName:                 "testing1.bak.1257894000",
				timeToSave:                604800,
				filter:                    filters,
			},
			newImageID: "ami-123456d",
			awsImages: []*ec2.Image{
				{
					ImageId:      aws.String("ami-123456a"),
					Name:         aws.String("testing1.bak.848590424"),
					CreationDate: getTimeSecondsBeforeNowAsString(604799),
				},
				{
					ImageId:      aws.String("ami-123456b"),
					Name:         aws.String("testing1.bak.438208309884"),
					CreationDate: getTimeSecondsBeforeNowAsString(604801),
				},
				{
					ImageId:      aws.String("ami-123456c"),
					Name:         aws.String("testing1.bak.4284932088"),
					CreationDate: getTimeSecondsBeforeNowAsString(604799),
				},
				{
					ImageId:      aws.String("ami-123456d"),
					Name:         aws.String("testing1.bak.993948322"),
					CreationDate: getTimeSecondsBeforeNowAsString(2),
				},
				{
					ImageId:      aws.String("ami-123456e"),
					Name:         aws.String("testing1.bak.3898349383"),
					CreationDate: getTimeSecondsBeforeNowAsString(1604802),
				},
			},
			deletes: []int{1},
			AWSSnapshots: []*ec2.Snapshot{
				{
					SnapshotId: aws.String("test1"),
					Description: aws.String(
						"This snapshot is taken from ami-123456b",
					),
				},
				{
					SnapshotId: aws.String("test2"),
					Description: aws.String(
						"This snapshot is taken from ami-123456e",
					),
				},
			},
		},
	}

	for _, test := range deregisterImageErrorTest {
		mockEC2iface.EXPECT().DescribeImages(
			&ec2.DescribeImagesInput{Filters: filters},
		).Return(
			&ec2.DescribeImagesOutput{Images: test.awsImages},
			nil,
		)
		for _, i := range test.deletes {
			mockEC2iface.EXPECT().DeregisterImage(
				&ec2.DeregisterImageInput{
					ImageId: test.awsImages[i].ImageId,
					DryRun:  aws.Bool(false),
				},
			).Return(
				nil,
				errors.New("Another error blah blah"),
			)
		}
		err := test.s.removeOldImage(test.newImageID)
		if err == nil {
			t.Error("Expect and error but got nil")
		}
	}

	var deleteSnapshotByDescriptionErrorTests = []struct {
		s            *svcEC2
		newImageID   string
		awsImages    []*ec2.Image
		deletes      []int
		AWSSnapshots []*ec2.Snapshot
	}{
		{
			s: &svcEC2{
				svc: mockEC2iface,
				imageNameWithoutTimestamp: "testing1.bak",
				imageName:                 "testing1.bak.1257894000",
				timeToSave:                604800,
				filter:                    filters,
			},
			newImageID: "ami-123456d",
			awsImages: []*ec2.Image{
				{
					ImageId:      aws.String("ami-123456a"),
					Name:         aws.String("testing1.bak.848590424"),
					CreationDate: getTimeSecondsBeforeNowAsString(604799),
				},
				{
					ImageId:      aws.String("ami-123456b"),
					Name:         aws.String("testing1.bak.438208309884"),
					CreationDate: getTimeSecondsBeforeNowAsString(604801),
				},
				{
					ImageId:      aws.String("ami-123456c"),
					Name:         aws.String("testing1.bak.4284932088"),
					CreationDate: getTimeSecondsBeforeNowAsString(604799),
				},
				{
					ImageId:      aws.String("ami-123456d"),
					Name:         aws.String("testing1.bak.993948322"),
					CreationDate: getTimeSecondsBeforeNowAsString(2),
				},
				{
					ImageId:      aws.String("ami-123456e"),
					Name:         aws.String("testing1.bak.3898349383"),
					CreationDate: getTimeSecondsBeforeNowAsString(1604802),
				},
			},
			deletes: []int{1},
			AWSSnapshots: []*ec2.Snapshot{
				{
					SnapshotId: aws.String("test1"),
					Description: aws.String(
						"This snapshot is taken from ami-123456b",
					),
				},
			},
		},
	}

	for _, test := range deleteSnapshotByDescriptionErrorTests {
		mockEC2iface.EXPECT().DescribeImages(
			&ec2.DescribeImagesInput{Filters: filters},
		).Return(
			&ec2.DescribeImagesOutput{Images: test.awsImages},
			nil,
		)
		for index, i := range test.deletes {
			mockEC2iface.EXPECT().DeregisterImage(
				&ec2.DeregisterImageInput{
					ImageId: test.awsImages[i].ImageId,
					DryRun:  aws.Bool(false),
				},
			).Return(
				&ec2.DeregisterImageOutput{},
				nil,
			)
			mockEC2iface.EXPECT().DescribeSnapshots(
				&ec2.DescribeSnapshotsInput{
					Filters: filters,
				},
			).Return(
				&ec2.DescribeSnapshotsOutput{
					Snapshots: test.AWSSnapshots,
				},
				nil,
			)
			mockEC2iface.EXPECT().DeleteSnapshot(
				&ec2.DeleteSnapshotInput{
					SnapshotId: test.AWSSnapshots[index].SnapshotId,
					DryRun:     aws.Bool(false),
				},
			).Return(
				nil,
				errors.New("Some error blah blah blah"),
			)
		}
		err := test.s.removeOldImage(test.newImageID)
		if err == nil {
			t.Error("Expected an error but got nil")
		}
	}
}

func TestDeleteSnapshotByDescription(t *testing.T) {
	mockEC2iface, ctrl := getMocks(t)
	defer ctrl.Finish()

	var filters = []*ec2.Filter{
		{
			Name: aws.String("owner-id"),
			Values: []*string{
				aws.String("533779774295"),
			},
		},
	}

	var happyPathTests = []struct {
		s                  *svcEC2
		imageIdToDelete    string
		AWSSnapshots       []*ec2.Snapshot
		snapshotIdToDelete *string
	}{
		{
			s: &svcEC2{
				svc: mockEC2iface,
				imageNameWithoutTimestamp: "testing123.bak",
				imageName:                 "testing123.bak.1257894000",
				timeToSave:                604800,
				filter:                    filters,
			},
			imageIdToDelete: "img-1323949t",
			AWSSnapshots: []*ec2.Snapshot{
				{
					SnapshotId: aws.String("test1"),
					Description: aws.String(
						"This is snapshot taken from img-1323949t",
					),
				},
				{
					SnapshotId: aws.String("test2"),
					Description: aws.String(
						"This is snapshot taken from img-fdafkdlkfj",
					),
				},
			},
			snapshotIdToDelete: aws.String("test1"),
		},
		{
			s: &svcEC2{
				svc: mockEC2iface,
				imageNameWithoutTimestamp: "testing123.bak",
				imageName:                 "testing123.bak.1257894000",
				timeToSave:                604800,
				filter:                    filters,
			},
			imageIdToDelete: "img-1323949t",
			AWSSnapshots: []*ec2.Snapshot{
				{
					SnapshotId: aws.String("test1"),
					Description: aws.String(
						"This is snapshot taken from img-1323949f",
					),
				},
				{
					SnapshotId: aws.String("test2"),
					Description: aws.String(
						"This is snapshot taken from img-1323949t",
					),
				},
			},
			snapshotIdToDelete: aws.String("test2"),
		},
		{
			s: &svcEC2{
				svc: mockEC2iface,
				imageNameWithoutTimestamp: "testing123.bak",
				imageName:                 "testing123.bak.1257894000",
				timeToSave:                604800,
				filter:                    filters,
			},
			imageIdToDelete: "img-1323949t",
			AWSSnapshots: []*ec2.Snapshot{
				{
					SnapshotId: aws.String("test1"),
					Description: aws.String(
						"This is snapshot taken from img-fdjkfjkj",
					),
				},
				{
					SnapshotId: aws.String("test2"),
					Description: aws.String(
						"This is snapshot taken from img-fdfdabbvd",
					),
				},
			},
			snapshotIdToDelete: nil,
		},
		{
			s: &svcEC2{
				svc: mockEC2iface,
				imageNameWithoutTimestamp: "testing123.bak",
				imageName:                 "testing123.bak.1257894000",
				timeToSave:                604800,
				filter:                    filters,
			},
			imageIdToDelete:    "img-1323949t",
			AWSSnapshots:       []*ec2.Snapshot{},
			snapshotIdToDelete: nil,
		},
	}

	for _, test := range happyPathTests {
		mockEC2iface.EXPECT().DescribeSnapshots(
			&ec2.DescribeSnapshotsInput{
				Filters: filters,
			},
		).Return(
			&ec2.DescribeSnapshotsOutput{
				Snapshots: test.AWSSnapshots,
			},
			nil,
		)
		if test.snapshotIdToDelete != nil {
			mockEC2iface.EXPECT().DeleteSnapshot(
				&ec2.DeleteSnapshotInput{
					SnapshotId: test.snapshotIdToDelete,
					DryRun:     aws.Bool(false),
				},
			).Return(
				&ec2.DeleteSnapshotOutput{},
				nil,
			)
		}
		err := test.s.deleteSnapshotByDescription(test.imageIdToDelete)
		if err != nil {
			t.Errorf("Expected nil but got error")
		}
	}

	var validateNegativeTests = func(
		err error,
		expectedMsg string,
		imageName string,
	) {
		if err == nil {
			t.Errorf("Expected an error by got nil but got error")
		}
		e := deleteError{
			imageName: imageName,
			msg:       expectedMsg,
		}
		if err.Error() != e.Error() {
			t.Errorf(
				"Expected error message of '%s' but got '%s'",
				e.Error(),
				err.Error(),
			)
		}
	}

	var describeSnapshotsErrorTests = []struct {
		s               *svcEC2
		imageIdToDelete string
		AWSSnapshots    []*ec2.Snapshot
		awsErr          awserr.Error
	}{
		{
			s: &svcEC2{
				svc: mockEC2iface,
				imageNameWithoutTimestamp: "testing123.bak",
				imageName:                 "testing123.bak.1257894000",
				timeToSave:                604800,
				filter:                    filters,
			},
			imageIdToDelete: "img-1323949t",
			AWSSnapshots:    nil,
			awsErr: awserr.New(
				"01-01",
				"Something when wrong",
				nil,
			),
		},
	}

	for _, test := range describeSnapshotsErrorTests {
		mockEC2iface.EXPECT().DescribeSnapshots(
			&ec2.DescribeSnapshotsInput{
				Filters: []*ec2.Filter{
					{
						Name: aws.String("owner-id"),
						Values: []*string{
							aws.String("533779774295"),
						},
					},
				},
			},
		).Return(
			&ec2.DescribeSnapshotsOutput{},
			test.awsErr,
		)

		err := test.s.deleteSnapshotByDescription(test.imageIdToDelete)
		validateNegativeTests(
			err,
			fmt.Sprintf(
				"Could not get snapshot list for deletion with msg %s",
				test.awsErr.Error(),
			),
			test.s.imageName,
		)

		var deleteSnapshotErrorTests = []struct {
			s                  *svcEC2
			imageIdToDelete    string
			AWSSnapshots       []*ec2.Snapshot
			snapshotIdToDelete *string
			awsErr             awserr.Error
		}{
			{
				s: &svcEC2{
					svc: mockEC2iface,
					imageNameWithoutTimestamp: "testing123.bak",
					imageName:                 "testing123.bak.1257894000",
					timeToSave:                604800,
					filter:                    filters,
				},
				imageIdToDelete: "img-1323949t",
				AWSSnapshots: []*ec2.Snapshot{
					{
						SnapshotId: aws.String("test1"),
						Description: aws.String(
							"This is snapshot taken from img-1323949t",
						),
					},
					{
						SnapshotId: aws.String("test2"),
						Description: aws.String(
							"This is snapshot taken from img-fdafkdlkfj",
						),
					},
				},
				snapshotIdToDelete: aws.String("test1"),
				awsErr: awserr.New(
					"01-01",
					"SnapshotId does not Exist",
					nil,
				),
			},
			{
				s: &svcEC2{
					svc: mockEC2iface,
					imageNameWithoutTimestamp: "testing123.bak",
					imageName:                 "testing123.bak.1257894000",
					timeToSave:                604800,
					filter:                    filters,
				},
				imageIdToDelete: "img-1323949t",
				AWSSnapshots: []*ec2.Snapshot{
					{
						SnapshotId: aws.String("test1"),
						Description: aws.String(
							"This is snapshot taken from img-1323949f",
						),
					},
					{
						SnapshotId: aws.String("test2"),
						Description: aws.String(
							"This is snapshot taken from img-1323949t",
						),
					},
				},
				snapshotIdToDelete: aws.String("test2"),
				awsErr: awserr.New(
					"01-02",
					"Something when wrong",
					nil,
				),
			},
		}

		for _, test := range deleteSnapshotErrorTests {
			mockEC2iface.EXPECT().DescribeSnapshots(
				&ec2.DescribeSnapshotsInput{
					Filters: []*ec2.Filter{
						{
							Name: aws.String("owner-id"),
							Values: []*string{
								aws.String("533779774295"),
							},
						},
					},
				},
			).Return(
				&ec2.DescribeSnapshotsOutput{
					Snapshots: test.AWSSnapshots,
				},
				nil,
			)
			if test.snapshotIdToDelete != nil {
				mockEC2iface.EXPECT().DeleteSnapshot(
					&ec2.DeleteSnapshotInput{
						SnapshotId: test.snapshotIdToDelete,
						DryRun:     aws.Bool(false),
					},
				).Return(
					&ec2.DeleteSnapshotOutput{},
					test.awsErr,
				)
			}
			err := test.s.deleteSnapshotByDescription(test.imageIdToDelete)
			var imageName *string
			for _, snapshot := range test.AWSSnapshots {
				if *snapshot.SnapshotId == *test.snapshotIdToDelete {
					imageName = snapshot.Description
				}
			}
			validateNegativeTests(err, test.awsErr.Error(), *imageName)
		}
	}
}

func TestDeleteError(t *testing.T) {
	tests := []struct {
		imageName string
		msg       string
	}{
		{
			imageName: "pd.test.i.20150607",
			msg:       "Delete failed, image does not exist",
		},
		{
			imageName: "pd.test.i",
			msg:       "Describe Reservation did not work",
		},
	}

	for _, test := range tests {
		d := deleteError{
			test.imageName,
			test.msg,
		}
		expect := fmt.Sprintf(
			"Image delete failed for image %s with \"%s\"",
			test.imageName,
			test.msg,
		)
		if d.Error() != expect {
			t.Errorf("Expected %s got %s", expect, d.Error())
		}
	}
}

func TestCreateNameWithTimeStamp(t *testing.T) {
	var (
		tests = []string{
			"pd.test.1.bak",
			"Int_123442.BAK",
			"Adquire.image.BAK",
			"ami-49504nj9",
			"_234354566",
			"90)9900)0987.bak",
		}
		negativeTestsAfter = []string{
			"pd.test.1.bak",
			"int-123442.bak",
			"_1adquire.image.bak",
			"AMI-49504nj9",
		}
		negativeTestsBefore = []string{
			"pd.test.1.bak",
			"pss_1adfg2.bak",
			"adquire.image.bak",
			"ami-49504nj9",
		}
	)
	for _, test := range tests {
		expect := fmt.Sprintf(
			"%s.%s",
			test,
			time.Now().Format("20060102150405"),
		)
		result := createNameWithTimestamp(test)
		if result != expect {
			t.Errorf("Expected %s got %s", expect, result)
		}
	}
	for _, test := range negativeTestsAfter {
		notExpected := fmt.Sprintf(
			"%s.%s",
			test,
			time.Now().Add(
				time.Minute*time.Duration(rand.Int31n(100)),
			).Format("20060102150405"),
		)
		result := createNameWithTimestamp(test)
		if notExpected == result {
			t.Errorf("Did not expect %s and but got %s", notExpected, result)
		}
	}
	for _, test := range negativeTestsBefore {
		notExpected := fmt.Sprintf(
			"%s.%s",
			test,
			time.Now().Add(
				-time.Minute*time.Duration(rand.Int31n(100)),
			).Format("20060102150405"),
		)
		result := createNameWithTimestamp(test)
		if notExpected == result {
			t.Errorf("Did not expect %s and but got %s", notExpected, result)
		}
	}
}
