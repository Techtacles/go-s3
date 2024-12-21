package main

import (
	"bytes"
	"context"
	"fmt"
	"os"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

const (
	bucketName string = "test-golang-s3-techtacles-bucket"
	filename   string = "file.txt"
)

var (
	instanceId string
	err        error
)

func main() {
	// ec2_out, err := createEc2(context.Background(), "us-east-1")
	// if err != nil {
	// 	fmt.Printf("The error is %s", err)
	// }
	// fmt.Println(ec2_out)
	cfg, err := config.LoadDefaultConfig(context.Background(),
		config.WithRegion("us-east-1"),
		config.WithSharedConfigProfile("personal"),
	)
	if err != nil {
		panic("Error loading configuration")
	}
	s3_client := s3.NewFromConfig(cfg)
	create_bucket_out, err := create_s3_bucket(context.Background(), s3_client, bucketName)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(create_bucket_out)

	put_obj_out, err := upload_to_s3(context.Background(), s3_client, bucketName)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(put_obj_out)
	fmt.Println("Done creating and uploading to s3 bucket")

}

func create_s3_bucket(ctx context.Context, s3_client *s3.Client, bucket_name string) (string, error) {
	_, err = s3_client.CreateBucket(ctx, &s3.CreateBucketInput{
		Bucket: aws.String(bucket_name),
	})
	if err != nil {
		return "", fmt.Errorf("error creating s3 bucket %q", err)
	}

	return fmt.Sprintf("Successfully created bucket %s", bucket_name), nil

}

func upload_to_s3(ctx context.Context, s3_client *s3.Client, bucket_name string) (string, error) {
	file_byte, err := os.ReadFile(filename)
	if err != nil {
		return "", fmt.Errorf("unable to read file %q", err)

	}
	reader := bytes.NewReader(file_byte)
	_, err = s3_client.PutObject(ctx, &s3.PutObjectInput{
		Bucket: &bucket_name,
		Key:    aws.String("file.txt"),
		Body:   reader,
	})
	if err != nil {
		return "", fmt.Errorf("unable to put object to s3 %q", err)
	}
	return "Successfully put file to bucket", nil
}

func createEc2(ctx context.Context, region string) (string, error) {
	// create config. Common to all aws packages in go
	cfg, err := config.LoadDefaultConfig(ctx,
		config.WithRegion("us-east-1"),
		config.WithSharedConfigProfile("personal"),
	)

	if err != nil {
		return "", fmt.Errorf("unable to load config")
	}

	// create ec2 config
	ec2Client := ec2.NewFromConfig(cfg)
	// create key pair
	keypair_output, err := ec2Client.CreateKeyPair(ctx, &ec2.CreateKeyPairInput{
		KeyName: aws.String("go-aws-demo"), // KeyName accepts only pointer string. So aws.String helps us do that
	})
	if err != nil {
		return "", fmt.Errorf("unable to create key pair")
	}
	// Describe images
	image_output, err := ec2Client.DescribeImages(ctx, &ec2.DescribeImagesInput{
		Filters: []types.Filter{
			{
				Name:   aws.String("name"),
				Values: []string{"ubuntu/images/hvm-ssd/ubuntu-focal-20.04-amd64-server-*"},
			},
			{
				Name:   aws.String("virtualization-type"),
				Values: []string{"hvm"},
			},
		},
		Owners: []string{"099720109477"},
	})
	if err != nil {
		return "", fmt.Errorf("unable to describe images")
	}
	if len(image_output.Images) == 0 {
		return "", fmt.Errorf("image_output.Images is of 0 length")
	}
	// run instances
	instance, err := ec2Client.RunInstances(ctx, &ec2.RunInstancesInput{
		ImageId:      image_output.Images[0].ImageId,
		KeyName:      keypair_output.KeyName,
		MinCount:     aws.Int32(1),
		MaxCount:     aws.Int32(1),
		InstanceType: types.InstanceTypeT3Micro,
	})
	if err != nil {
		return "", fmt.Errorf("run instance error")
	}
	if len(instance.Instances) == 0 {
		return "", fmt.Errorf("instance is of 0 length")
	}
	return "Successfully launched ec2", nil

}
