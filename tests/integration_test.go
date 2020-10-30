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

	"golang.org/x/crypto/ssh"

	"github.com/aws/aws-sdk-go/aws"
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
	cleanupAWSResources()
	cleanup()
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

	resourcesTypesToRemove := []string{"Instance", "SecurityGroup", "RouteTable", "Subnet", "EIP", "InternetGateway", "NatGateway", "VPC"}

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
		default:
			log.Println("Wawaweewa.")
		}
	}

	log.Println("Removing resource group: ", rgName)
	rgDelInp := resourcegroups.DeleteGroupInput{
		GroupName: aws.String(rgName),
	}
	rgDelOut, rgDelErr := rgClient.DeleteGroup(&rgDelInp)
	log.Println("Output: ", rgDelOut)
	log.Println("Error: ", rgDelErr)

	log.Println("Removing key pair: ", kpName)
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
		log.Fatal("Cannot get session.")
	}

	ec2Client := ec2.New(newSession)

	log.Println("Removing...")

	for _, ec2ToRemove := range ec2sToRemove {

		ec2ToRemoveID := strings.Split(*ec2ToRemove.ResourceArn, "/")[1]
		log.Println("Removing instance with ID: ", ec2ToRemoveID)

		ec2DescInp := &ec2.DescribeInstancesInput{
			InstanceIds: []*string{&ec2ToRemoveID},
		}

		outDesc, errDesc := ec2Client.DescribeInstances(ec2DescInp)
		log.Printf("Desc-Output: %s,\nDesc-Error: %s", outDesc, errDesc)

		if outDesc.Reservations != nil {

			instancesToTerminateInp := &ec2.TerminateInstancesInput{
				InstanceIds: []*string{&ec2ToRemoveID},
			}

			output, err := ec2Client.TerminateInstances(instancesToTerminateInp)
			log.Printf("Terminate-Output: %s,\nTerminate-Error: %s", output, err)

			errWait := ec2Client.WaitUntilInstanceTerminated(ec2DescInp)
			log.Printf("Error: %s", errWait)
		}

	}

}

func removeRouteTables(rtsToRemove []*resourcegroups.ResourceIdentifier) {
	newSession, err := session.NewSession(&aws.Config{Region: aws.String("eu-central-1")})
	if err != nil {
		log.Fatal("Cannot get session.")
	}

	ec2Client := ec2.New(newSession)

	log.Println("Removing...")

	for _, rtToRemove := range rtsToRemove {
		rtIDToRemove := strings.Split(*rtToRemove.ResourceArn, "/")[1]
		log.Println("rtIDToRemove: ", rtIDToRemove)

		rtToDeleteInp := &ec2.DeleteRouteTableInput{
			RouteTableId: &rtIDToRemove,
		}

		output, err := ec2Client.DeleteRouteTable(rtToDeleteInp)
		log.Println("Output: ", output)
		log.Println("Error: ", err)
	}

}

func removeSecurityGroup(sgsToRemove []*resourcegroups.ResourceIdentifier) {
	newSession, err := session.NewSession(&aws.Config{Region: aws.String("eu-central-1")})
	if err != nil {
		log.Fatal("Cannot get session.")
	}

	ec2Client := ec2.New(newSession)

	log.Println("Removing...")

	for _, sgToRemove := range sgsToRemove {
		sgIDToRemove := strings.Split(*sgToRemove.ResourceArn, "/")[1]
		log.Println("sgIdToRemove: ", sgIDToRemove)

		secGrpInp := &ec2.DeleteSecurityGroupInput{GroupId: &sgIDToRemove}

		output, err := ec2Client.DeleteSecurityGroup(secGrpInp)
		log.Println("Output: ", output)
		log.Println("Error: ", err)
	}

}

func removeInternetGateway(igsToRemove []*resourcegroups.ResourceIdentifier) {
	newSession, err := session.NewSession(&aws.Config{Region: aws.String("eu-central-1")})
	if err != nil {
		log.Fatal("Cannot get session.")
	}

	ec2Client := ec2.New(newSession)

	log.Println("Removing...")

	for _, igToRemove := range igsToRemove {
		igIDToRemove := strings.Split(*igToRemove.ResourceArn, "/")[1]
		log.Println("igIdToRemove: ", igIDToRemove)

		igDescribeInp := &ec2.DescribeInternetGatewaysInput{
			InternetGatewayIds: []*string{&igIDToRemove},
		}

		descOut, descErr := ec2Client.DescribeInternetGateways(igDescribeInp)
		vpcID := *descOut.InternetGateways[0].Attachments[0].VpcId
		log.Println("Describe-IG-Output: ", descOut)
		log.Println("Describe-IG-Error: ", descErr)

		igDetachInp := &ec2.DetachInternetGatewayInput{
			InternetGatewayId: &igIDToRemove,
			VpcId:             &vpcID,
		}

		detachOut, detachErr := ec2Client.DetachInternetGateway(igDetachInp)
		log.Println("Detach-IG-Output: ", detachOut)
		log.Println("Detach-IG-Error: ", detachErr)

		igDeleteInp := &ec2.DeleteInternetGatewayInput{
			InternetGatewayId: &igIDToRemove,
		}

		delOut, delErr := ec2Client.DeleteInternetGateway(igDeleteInp)
		log.Println("Delete-IG-Output: ", delOut)
		log.Println("Delete-IG-Error: ", delErr)

	}
}

func removeNatGateway(ngsToRemove []*resourcegroups.ResourceIdentifier) {
	newSession, err := session.NewSession(&aws.Config{Region: aws.String("eu-central-1")})
	if err != nil {
		log.Fatal("Cannot get session.")
	}

	ec2Client := ec2.New(newSession)

	log.Println("Removing...")

	for _, ngToRemove := range ngsToRemove {
		ngIDToRemove := strings.Split(*ngToRemove.ResourceArn, "/")[1]
		log.Println("ngIdToRemove: ", ngIDToRemove)

		ngInp := &ec2.DeleteNatGatewayInput{
			NatGatewayId: &ngIDToRemove,
		}

		output, err := ec2Client.DeleteNatGateway(ngInp)
		log.Println("Output: ", output)
		log.Println("Error: ", err)
	}

}

func removeSubnet(subnetsToRemove []*resourcegroups.ResourceIdentifier) {
	newSession, err := session.NewSession(&aws.Config{Region: aws.String("eu-central-1")})
	if err != nil {
		log.Fatal("Cannot get session.")
	}

	ec2Client := ec2.New(newSession)

	log.Println("Removing...")

	for _, subnetToRemove := range subnetsToRemove {
		subnetIDToRemove := strings.Split(*subnetToRemove.ResourceArn, "/")[1]
		log.Println("subnetIdToRemove: ", subnetIDToRemove)

		subnetInp := &ec2.DeleteSubnetInput{
			SubnetId: &subnetIDToRemove,
		}

		output, err := ec2Client.DeleteSubnet(subnetInp)
		log.Println("Output: ", output)
		log.Println("Error: ", err)
	}

}

func removeVpc(vpcsToRemove []*resourcegroups.ResourceIdentifier) {
	newSession, err := session.NewSession(&aws.Config{Region: aws.String("eu-central-1")})
	if err != nil {
		log.Fatal("Cannot get session.")
	}

	ec2Client := ec2.New(newSession)

	log.Println("Removing...")

	for _, vpcToRemove := range vpcsToRemove {
		vpcIDToRemove := strings.Split(*vpcToRemove.ResourceArn, "/")[1]
		log.Println("vpcIdToRemove: ", vpcIDToRemove)

		vpcToDeleteInp := &ec2.DeleteVpcInput{
			VpcId: &vpcIDToRemove,
		}

		output, err := ec2Client.DeleteVpc(vpcToDeleteInp)
		log.Println("Output: ", output)
		log.Println("Error: ", err)
	}
}

func removeKeyPair(kpName string) {
	newSession, err := session.NewSession(&aws.Config{Region: aws.String("eu-central-1")})
	if err != nil {
		log.Fatal("Cannot get session.", err)
	}

	ec2Client := ec2.New(newSession)

	log.Println("Removing...")

	removeKeyInp := &ec2.DeleteKeyPairInput{
		KeyName: &kpName,
	}

	output, err := ec2Client.DeleteKeyPair(removeKeyInp)
	log.Println("Output: ", output)
	log.Println("Error: ", err)
}

func releaseAddresses(eipName string) {
	newSession, err := session.NewSession(&aws.Config{Region: aws.String("eu-central-1")})
	if err != nil {
		log.Fatal("Cannot get session.", err)
	}

	ec2Client := ec2.New(newSession)

	log.Println("Removing...")

	describeEips, err := ec2Client.DescribeAddresses(&ec2.DescribeAddressesInput{})
	if err != nil {
		log.Fatal("Cannot get EIP list.")
	}

	for _, eip := range describeEips.Addresses {

		tagPresent := false
		if eip.Tags != nil {
			tagPresent = checkIfTagPresent(eipName, eip.Tags)
			if tagPresent {
				log.Printf("Remove EIP with AllocationId: %s", *eip.AllocationId)

				eipToReleaseInp := &ec2.ReleaseAddressInput{
					AllocationId: eip.AllocationId,
				}

				output, err := ec2Client.ReleaseAddress(eipToReleaseInp)
				log.Println("Output: ", output)
				log.Println("Error: ", err)
			}
		}

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
