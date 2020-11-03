package main

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
	"time"

	"golang.org/x/crypto/ssh"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/resourcegroups"
)

const (
	imageTag   = "epiphanyplatform/awsbi:0.0.1"
	awsTag     = "awsbi-module"
	moduleName = "awsbi-module"
	sshKeyName = "vms_rsa"
)

var (
	awsAccessKey              string
	awsSecretKey              string
	sharedAbsoluteFilePath, _ = filepath.Abs("./shared")
	mountDir                  = sharedAbsoluteFilePath + ":/shared"
	dockerExecPath, _         = exec.LookPath("docker")
)

func TestMain(m *testing.M) {
	cleanup()
	cleanupAWSResources()
	setup()
	log.Println("Run tests")
	exitVal := m.Run()
	log.Println("Finish test")
	cleanup()
	cleanupAWSResources()
	os.Exit(exitVal)
}

func TestInitShouldCreateProperFileAndFolder(t *testing.T) {
	// given
	stateFilePath := "shared/state.yml"
	expectedFileContentRegexp := "kind: state\nawsbi:\n  status: initialized"

	// when
	runCommand(dockerExecPath, "run", "--rm", "-v", mountDir, "-t", imageTag, "init", "M_NAME="+moduleName)

	data, err := ioutil.ReadFile(stateFilePath)

	if err != nil {
		t.Fatal("Cannot read state file: ", stateFilePath)
	}

	fileContent := string(data)

	matched, _ := regexp.MatchString(expectedFileContentRegexp, fileContent)

	// then
	if !matched {
		t.Error("Expected to find expression matching:\n", expectedFileContentRegexp, "\nbut found:\n", fileContent)
	}

}

func TestOnPlanWithDefaultsShouldDisplayPlan(t *testing.T) {
	// given
	var stdout, stderr bytes.Buffer

	expectedOutputRegexp := ".*Plan: 14 to add, 0 to change, 0 to destroy.*"

	// when
	stdout, stderr = runCommand(dockerExecPath, "run", "--rm", "-v", mountDir, "-t", imageTag, "plan", awsAccessKey, awsSecretKey)

	if stderr.Len() > 0 {
		t.Fatal("There was an error during executing a command. ", string(stderr.Bytes()))
	}

	outStr := string(stdout.Bytes())

	matched, _ := regexp.MatchString(expectedOutputRegexp, outStr)

	// then
	if !matched {
		t.Error("Expected to find expression matching:\n", expectedOutputRegexp, "\nbut found:\n", outStr)
	}

}

func TestOnApplyShouldCreateEnvironment(t *testing.T) {
	// given
	var stdout, stderr bytes.Buffer
	expectedOutputRegexp := ".*Apply complete! Resources: 14 added, 0 changed, 0 destroyed.*"

	// when
	stdout, stderr = runCommand(dockerExecPath, "run", "--rm", "-v", mountDir, "-t", imageTag, "apply", awsAccessKey, awsSecretKey)

	if stderr.Len() > 0 {
		t.Fatal("There was an error during executing a command. ", string(stderr.Bytes()))
	}

	outStr := string(stdout.Bytes())

	matched, _ := regexp.MatchString(expectedOutputRegexp, outStr)

	// then
	if !matched {
		t.Error("Expected to find expression matching:\n", expectedOutputRegexp, "\nbut found:\n", outStr)
	}

}

func TestShouldCheckNumberOfVms(t *testing.T) {
	// given
	instancesNumber := 1

	newSession, err := session.NewSession(&aws.Config{Region: aws.String("eu-central-1")})
	if err != nil {
		t.Fatal("Cannot get session.")
	}

	// when

	// ec2
	ec2Client := ec2.New(newSession)

	ec2DescInp := &ec2.DescribeInstancesInput{
		Filters: []*ec2.Filter{
			{
				Name: aws.String("tag:Name"),
				Values: []*string{
					aws.String(awsTag),
				},
			},
		},
	}

	ec2Result, err := ec2Client.DescribeInstances(ec2DescInp)

	// rg
	// rgClient := resourcegroups.New(newSession)

	// rgName := "rg-" + moduleName

	// rgResourcesList, err := rgClient.ListGroupResources(&resourcegroups.ListGroupResourcesInput{
	// 	GroupName: aws.String(rgName),
	// })

	// then
	if err != nil {
		t.Fatal("There was an error. ", err)
	}

	if len(ec2Result.Reservations[0].Instances) != 1 {
		t.Error("Expected ", instancesNumber, "instance, got ", len(ec2Result.Reservations[0].Instances))
	}

}

// func TestOnDestroyPlanShouldDisplayDestroyPlan(t *testing.T) {
// 	// given
// 	var stdout, stderr bytes.Buffer
// 	expectedOutputRegexp := "Plan: 0 to add, 0 to change, 14 to destroy"

// 	// when
// 	stdout, stderr = runCommand(dockerExecPath, "run", "--rm", "-v", mountDir, "-t", imageTag, "plan-destroy", awsAccessKey, awsSecretKey)

// 	if stderr.Len() > 0 {
// 		t.Fatal("There was an error during executing a command. ", string(stderr.Bytes()))
// 	}

// 	outStr := string(stdout.Bytes())

// 	matched, _ := regexp.MatchString(expectedOutputRegexp, outStr)

// 	// then
// 	if !matched {
// 		t.Error("Expected to find expression matching:\n", expectedOutputRegexp, "\nbut found:\n", outStr)
// 	}
// }

// func TestOnDestroyShouldDestroyEnvironment(t *testing.T) {
// 	// given
// 	var stdout, stderr bytes.Buffer

// 	expectedOutputRegexp := "Apply complete! Resources: 0 added, 0 changed, 14 destroyed."

// 	// when
// 	stdout, stderr = runCommand(dockerExecPath, "run", "--rm", "-v", mountDir, "-t", imageTag, "destroy", awsAccessKey, awsSecretKey)

// 	if stderr.Len() > 0 {
// 		t.Fatal("There was an error during executing a command. ", string(stderr.Bytes()))
// 	}

// 	outStr := string(stdout.Bytes())

// 	matched, _ := regexp.MatchString(expectedOutputRegexp, outStr)

// 	// then
// 	if !matched {
// 		t.Error("Expected to find expression matching:\n", expectedOutputRegexp, "\nbut found:\n", outStr)
// 	}
// }

func setup() {
	awsAccessKey = os.Getenv("AWS_ACCESS_KEY_ID")
	log.Println("Initialize test")
	if len(awsAccessKey) == 0 {
		log.Fatalf("expected non-empty AWS_ACCESS_KEY_ID environment variable")
	}
	awsAccessKey = "M_AWS_ACCESS_KEY=" + awsAccessKey

	awsSecretKey = os.Getenv("AWS_SECRET_ACCESS_KEY")
	if len(awsSecretKey) == 0 {
		log.Fatalf("expected non-empty AWS_SECRET_ACCESS_KEY environment variable")
	}
	awsSecretKey = "M_AWS_SECRET_KEY=" + awsSecretKey

	err := os.MkdirAll(sharedAbsoluteFilePath, os.ModePerm)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("Generating Keys")
	err = generateRsaKeyPair(sharedAbsoluteFilePath, sshKeyName)
	if err != nil {
		log.Fatal(err)
	}
}

func cleanup() {
	log.Println("Starting cleanup.")
	err := os.RemoveAll(sharedAbsoluteFilePath)
	if err != nil {
		log.Fatal("Cannot remove data folder. ", err)
	}
	log.Println("Cleanup finished.")
}

func cleanupAWSResources() {

	session, err := session.NewSession(&aws.Config{Region: aws.String("eu-central-1")})
	if err != nil {
		log.Fatal("Cannot get session.")
	}

	rgClient := resourcegroups.New(session)

	rgName := "rg-" + moduleName
	kpName := "kp-" + moduleName
	eipName := "eip-" + moduleName

	rgResourcesList, err := rgClient.ListGroupResources(&resourcegroups.ListGroupResourcesInput{
		GroupName: aws.String(rgName),
	})

	log.Println(rgResourcesList)

	resourcesTypesToRemove := []string{"Instance", "SecurityGroup", "NatGateway", "EIP", "InternetGateway", "Subnet", "RouteTable", "VPC"}

	for _, resourcesTypeToRemove := range resourcesTypesToRemove {

		filtered := make([]*resourcegroups.ResourceIdentifier, 0)
		for _, element := range rgResourcesList.ResourceIdentifiers {
			s := strings.Split(*element.ResourceType, ":")
			if s[4] == resourcesTypeToRemove {
				filtered = append(filtered, element)
			}

		}

		switch resourcesTypeToRemove {
		case "Instance":
			log.Println("Instance.")
			removeEc2s(filtered)
		case "EIP":
			log.Println("Releasing public EIPs.")
			releaseAddresses(eipName)
		case "RouteTable":
			log.Println("RouteTable.")
			removeRouteTables(filtered)
		case "InternetGateway":
			log.Println("InternetGateway.")
			removeInternetGateway(filtered)
		case "SecurityGroup":
			log.Println("SecurityGroup.")
			removeSecurityGroup(filtered)
		case "NatGateway":
			log.Println("NatGateway.")
			removeNatGateway(filtered)
		case "Subnet":
			log.Println("Subnet.")
			removeSubnet(filtered)
		case "VPC":
			log.Println("VPC.")
			removeVpc(filtered)
		}
	}

	removeResourceGroup(rgName)

	removeKeyPair(kpName)

}

func runCommand(commandWithParams ...string) (bytes.Buffer, bytes.Buffer) {
	var stdout, stderr bytes.Buffer
	dockerRunInit := &exec.Cmd{
		Path:   commandWithParams[0],
		Args:   commandWithParams,
		Stdout: &stdout,
		Stderr: &stderr,
	}
	if err := dockerRunInit.Run(); err != nil {
		log.Println("Error:", err)
	}

	return stdout, stderr
}

func generateRsaKeyPair(directory, name string) error {
	privateRsaKey, err := rsa.GenerateKey(rand.Reader, 4096)
	if err != nil {
		return err
	}
	pemBlock := &pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(privateRsaKey)}
	privateKeyBytes := pem.EncodeToMemory(pemBlock)

	publicRsaKey, err := ssh.NewPublicKey(&privateRsaKey.PublicKey)
	if err != nil {
		return err
	}
	publicKeyBytes := ssh.MarshalAuthorizedKey(publicRsaKey)

	err = ioutil.WriteFile(path.Join(directory, name), privateKeyBytes, 0600)
	if err != nil {
		return err
	}
	return ioutil.WriteFile(path.Join(directory, name+".pub"), publicKeyBytes, 0644)
}

func removeEc2s(ec2sToRemove []*resourcegroups.ResourceIdentifier) {
	newSession, err := session.NewSession(&aws.Config{Region: aws.String("eu-central-1")})
	if err != nil {
		log.Fatal("EC2: Cannot get session.")
	}

	ec2Client := ec2.New(newSession)

	for _, ec2ToRemove := range ec2sToRemove {

		ec2ToRemoveID := strings.Split(*ec2ToRemove.ResourceArn, "/")[1]
		log.Println("EC2: Removing instance with ID: ", ec2ToRemoveID)

		ec2DescInp := &ec2.DescribeInstancesInput{
			InstanceIds: []*string{&ec2ToRemoveID},
		}

		outDesc, errDesc := ec2Client.DescribeInstances(ec2DescInp)
		if errDesc != nil {
			log.Fatalf("EC2: Describe error: %s", errDesc)
		}
		log.Printf("EC2: Describe output: %s", outDesc)

		if outDesc.Reservations != nil {

			instancesToTerminateInp := &ec2.TerminateInstancesInput{
				InstanceIds: []*string{&ec2ToRemoveID},
			}

			outputTerm, errTerm := ec2Client.TerminateInstances(instancesToTerminateInp)
			if errTerm != nil {
				log.Fatalf("EC2: Terminate error: %s", outputTerm)
			}
			log.Printf("EC2: Terminate output: %s", outputTerm)

			errWait := ec2Client.WaitUntilInstanceTerminated(ec2DescInp)
			if errWait != nil {
				log.Fatalf("EC2: Waiting for termination error: %s", errWait)
			}
		}

	}

}

func removeRouteTables(rtsToRemove []*resourcegroups.ResourceIdentifier) {
	newSession, err := session.NewSession(&aws.Config{Region: aws.String("eu-central-1")})
	if err != nil {
		log.Fatal("RouteTable: Cannot get session.")
	}

	ec2Client := ec2.New(newSession)

	for _, rtToRemove := range rtsToRemove {
		rtIDToRemove := strings.Split(*rtToRemove.ResourceArn, "/")[1]
		log.Println("RouteTable: rtIDToRemove: ", rtIDToRemove)

		rtToDeleteInp := &ec2.DeleteRouteTableInput{
			RouteTableId: &rtIDToRemove,
		}

		output, err := ec2Client.DeleteRouteTable(rtToDeleteInp)

		if err != nil {
			log.Fatal("RouteTable: Deleting route table error: ", err)
		}

		log.Println("RouteTable: Deleting route table: ", output)
	}

}

func removeSecurityGroup(sgsToRemove []*resourcegroups.ResourceIdentifier) {
	newSession, err := session.NewSession(&aws.Config{Region: aws.String("eu-central-1")})
	if err != nil {
		log.Fatal("Security Group: Cannot get session.")
	}

	ec2Client := ec2.New(newSession)

	for _, sgToRemove := range sgsToRemove {
		sgIDToRemove := strings.Split(*sgToRemove.ResourceArn, "/")[1]
		log.Println("Security Group: sgIdToRemove: ", sgIDToRemove)

		secGrpInp := &ec2.DeleteSecurityGroupInput{GroupId: &sgIDToRemove}

		output, err := ec2Client.DeleteSecurityGroup(secGrpInp)
		if err != nil {
			log.Fatal("Security Group: Deleting security group error: ", err)
		}

		log.Println("Security Group: Deleting security group: ", output)
	}

}

func removeInternetGateway(igsToRemove []*resourcegroups.ResourceIdentifier) {
	newSession, err := session.NewSession(&aws.Config{Region: aws.String("eu-central-1")})
	if err != nil {
		log.Fatal("Internet Gateway: Cannot get session.")
	}

	ec2Client := ec2.New(newSession)

	for _, igToRemove := range igsToRemove {
		igIDToRemove := strings.Split(*igToRemove.ResourceArn, "/")[1]
		log.Println("Internet Gateway: igIdToRemove: ", igIDToRemove)

		igDescribeInp := &ec2.DescribeInternetGatewaysInput{
			InternetGatewayIds: []*string{&igIDToRemove},
		}

		descOut, descErr := ec2Client.DescribeInternetGateways(igDescribeInp)

		if descErr != nil {
			log.Fatal("Internet Gateway: Describing internet gateway error: ", descErr)
		}
		log.Println("Internet Gateway: Describing internet gateway: ", descOut)
		vpcID := *descOut.InternetGateways[0].Attachments[0].VpcId

		igDetachInp := &ec2.DetachInternetGatewayInput{
			InternetGatewayId: &igIDToRemove,
			VpcId:             &vpcID,
		}

		detachOut, detachErr := ec2Client.DetachInternetGateway(igDetachInp)
		if detachErr != nil {
			log.Fatal("Internet Gateway: Detaching internet gateway error: ", detachErr)
		}
		log.Println("Internet Gateway: Detaching internet gateway: ", detachOut)

		igDeleteInp := &ec2.DeleteInternetGatewayInput{
			InternetGatewayId: &igIDToRemove,
		}

		delOut, delErr := ec2Client.DeleteInternetGateway(igDeleteInp)
		if delErr != nil {
			log.Fatal("Internet Gateway: Deleting internet gateway error: ", delErr)
		}
		log.Println("Internet Gateway: Deleting internet gateway: ", delOut)
	}
}

func removeNatGateway(ngsToRemove []*resourcegroups.ResourceIdentifier) {
	newSession, err := session.NewSession(&aws.Config{Region: aws.String("eu-central-1")})
	if err != nil {
		log.Fatal("Nat Gateway: Cannot get session.")
	}

	ec2Client := ec2.New(newSession)

	for _, ngToRemove := range ngsToRemove {
		ngIDToRemove := strings.Split(*ngToRemove.ResourceArn, "/")[1]
		log.Println("Nat Gateway: ngIdToRemove: ", ngIDToRemove)

		descInp := &ec2.DescribeNatGatewaysInput{
			NatGatewayIds: []*string{&ngIDToRemove},
		}

		outDesc, errDesc := ec2Client.DescribeNatGateways(descInp)
		if errDesc != nil {

			if aerr, ok := errDesc.(awserr.Error); ok {
				if aerr.Code() != "NatGatewayNotFound" {
					log.Fatal("Nat Gateway: Describe error: ", errDesc)
				}
			}
		}
		log.Printf("Nat Gateway: Describe output: %s", outDesc)

		if outDesc.NatGateways != nil && *outDesc.NatGateways[0].State != "deleted" {
			ngInp := &ec2.DeleteNatGatewayInput{
				NatGatewayId: &ngIDToRemove,
			}

			output, err := ec2Client.DeleteNatGateway(ngInp)
			if err != nil {
				log.Fatal("Nat Gateway: Deleting NAT Gateway error: ", err)
			}
			log.Println("Nat Gateway: Deleting NAT Gateway: ", output)

			errWait := ec2Client.WaitUntilNatGatewayAvailable(descInp)
			if errWait != nil {
				if aerr, ok := errDesc.(awserr.Error); ok {
					if aerr.Code() != "ResourceNotReady" {
						log.Fatal("Nat Gateway: Wait error: ", errDesc)
					}
				}
			}
		}

	}

}

func removeSubnet(subnetsToRemove []*resourcegroups.ResourceIdentifier) {
	newSession, err := session.NewSession(&aws.Config{Region: aws.String("eu-central-1")})
	if err != nil {
		log.Fatal("Subnet: Cannot get session.")
	}

	ec2Client := ec2.New(newSession)

	for _, subnetToRemove := range subnetsToRemove {
		subnetIDToRemove := strings.Split(*subnetToRemove.ResourceArn, "/")[1]
		log.Println("Subnet: subnetIdToRemove: ", subnetIDToRemove)

		subnetInp := &ec2.DeleteSubnetInput{
			SubnetId: &subnetIDToRemove,
		}

		output, err := ec2Client.DeleteSubnet(subnetInp)
		if err != nil {
			log.Fatal("Subnet: Deleting subnet error: ", err)
		}
		log.Println("Subnet: Deleting subnet: ", output)
	}

}

func removeVpc(vpcsToRemove []*resourcegroups.ResourceIdentifier) {
	newSession, err := session.NewSession(&aws.Config{Region: aws.String("eu-central-1")})
	if err != nil {
		log.Fatal("VPC: Cannot get session.")
	}

	ec2Client := ec2.New(newSession)

	for _, vpcToRemove := range vpcsToRemove {
		vpcIDToRemove := strings.Split(*vpcToRemove.ResourceArn, "/")[1]
		log.Println("VPC: vpcIdToRemove: ", vpcIDToRemove)

		vpcToDeleteInp := &ec2.DeleteVpcInput{
			VpcId: &vpcIDToRemove,
		}

		output, err := ec2Client.DeleteVpc(vpcToDeleteInp)
		if err != nil {
			log.Println("VPC: Delete VPC error: ", err)
		}
		log.Println("VPC: Delete VPC: ", output)
	}
}

func removeKeyPair(kpName string) {
	newSession, err := session.NewSession(&aws.Config{Region: aws.String("eu-central-1")})
	if err != nil {
		log.Fatal("Key Pair: Cannot get session.", err)
	}

	ec2Client := ec2.New(newSession)

	removeKeyInp := &ec2.DeleteKeyPairInput{
		KeyName: &kpName,
	}

	output, err := ec2Client.DeleteKeyPair(removeKeyInp)
	if err != nil {
		log.Fatal("Key Pair: Deleting key pair error: ", err)
	}
	log.Println("Key Pair: Deleting key pair: ", output)
}

func releaseAddresses(eipName string) {
	newSession, err := session.NewSession(&aws.Config{Region: aws.String("eu-central-1")})
	if err != nil {
		log.Fatal("EIP: Cannot get session.", err)
	}

	ec2Client := ec2.New(newSession)

	describeEips, err := ec2Client.DescribeAddresses(&ec2.DescribeAddressesInput{})
	if err != nil {
		log.Fatal("EIP: Cannot get EIP list.", err)
	}

	for _, eip := range describeEips.Addresses {

		tagPresent := false
		if eip.Tags != nil {
			tagPresent = checkIfTagPresent(eipName, eip.Tags)
			if tagPresent {
				log.Printf("EIP: Releasing EIP with AllocationId: %s", *eip.AllocationId)

				eipToReleaseInp := &ec2.ReleaseAddressInput{
					AllocationId: eip.AllocationId,
				}

				found := true

				for retry := 0; retry <= 30 && found; retry++ {
					_, err := ec2Client.ReleaseAddress(eipToReleaseInp)
					if err != nil {
						if aerr, ok := err.(awserr.Error); ok {
							if aerr.Code() == "InvalidAllocationID.NotFound" {
								log.Print("EIP: Element not found.", err)
								found = false
								continue
							}
							if aerr.Code() != "AuthFailure" && aerr.Code() != "InvalidAllocationID.NotFound" {
								log.Fatal("EIP: Releasing EIP error: ", err)
							}
						}
					}
					log.Println("EIP: Releasing EIP. Retry: ", retry)
					time.Sleep(5 * time.Second)
				}
			}
		}

	}

}

func removeResourceGroup(rgToRemoveName string) {

	session, err := session.NewSession(&aws.Config{Region: aws.String("eu-central-1")})
	if err != nil {
		log.Fatal("Resource Group: Cannot get session.")
	}

	rgClient := resourcegroups.New(session)

	log.Println("Resource Group: Removing resource group: ", rgToRemoveName)
	rgDelInp := resourcegroups.DeleteGroupInput{
		GroupName: aws.String(rgToRemoveName),
	}
	rgDelOut, rgDelErr := rgClient.DeleteGroup(&rgDelInp)
	if rgDelErr != nil {
		if aerr, ok := rgDelErr.(awserr.Error); ok {
			if aerr.Code() == "NotFoundException" {
				log.Println("Resource Group: Resource group not found. ")
			} else {
				log.Fatal("Resource Group: Deleting resource group error: ", rgDelErr)
			}
		}

	} else {
		log.Println("Resource Group: Deleting resource group: ", rgDelOut)
	}

}

func checkIfTagPresent(toFind string, tags []*ec2.Tag) bool {

	for _, tag := range tags {
		if *tag.Value == toFind {
			return true
		}
	}
	return false
}
