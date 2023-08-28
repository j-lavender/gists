// This script runs a Terminal User Interface displaying a list of 
// EC2 Instances located in an AWS account. A valid set of AWS 
// credentials must be stored in ~/.aws/credentials before usage.
// After running, press 'h' to view the help menu.
// 
// Dependencies:
// 		go mod init aws-helpers
// 		go get .
// 
// Use:
// 		go run instances.go --profile default --region us-east-1
// 
// Build:
// 		go build instances.go
// 		./instances -p default -r us-east-1


package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/rivo/tview"
	"github.com/gdamore/tcell/v2"
)

var (
	profile 			string
	region  			string
	currentInstances 	[]*ec2.Instance
	app              	= tview.NewApplication()
	flex             	*tview.Flex
)

// Helper function to get the Name tag for an EC2 instance.
func getInstanceName(instance *ec2.Instance) string {
    for _, tag := range instance.Tags {
        if aws.StringValue(tag.Key) == "Name" && tag.Value != nil && *tag.Value != "" {
            return *tag.Value
        }
    }
    return *instance.InstanceId // default to instance ID if no Name tag found
}

// Helper function to filter the displayed instances by filtered state.
func fetchAndDisplayInstances(state string, svc *ec2.EC2, list *tview.List, detailsTable *tview.Table) {
	// Init params struct to pass to AWS SDK.
	params := &ec2.DescribeInstancesInput{}

	// Validate provided state.
    validStates := map[string]bool{
        "running":    true,
        "stopped":    true,
        "terminated": true,
    }

	// If a valid state is provided, then set the filter, else fetch all instances.
	// If not, no filter is set and all instances will be fetched.
    if validStates[state] {
        params.Filters = []*ec2.Filter{
            {
                Name:   aws.String("instance-state-name"),
                Values: []*string{aws.String(state)},
            },
        }
    }

	// Log error if API call to fetch instances fails.
    resp, err := svc.DescribeInstances(params)
    if err != nil {
        log.Fatalf("Failed to retrieve EC2 instances: %s", err)
    }

	// Clear the current list and set currentInstances slice to zero.
    list.Clear()
	currentInstances = nil

	// Use filter state to fetch instance list and append to currentInstances slice.
    for _, res := range resp.Reservations {
        for _, instance := range res.Instances {
            currentInstances = append(currentInstances, instance)
            list.AddItem(getInstanceName(instance), "", 0, nil)
        }
    }

	// If no instances exist for filter, clear the instanceDetails table. Otherwise display details for the first instance found.
	if len(currentInstances) == 0 {
        detailsTable.Clear()
        detailsTable.SetCell(0, 0, tview.NewTableCell("No instance details to display").SetAlign(tview.AlignCenter))
    } else {
        displayInstanceDetails(detailsTable, currentInstances[0])
    }
}

// Helper function to display the instance details of the first instance in the list. Function executes when application first loads.
func initialDisplay(svc *ec2.EC2, list *tview.List, detailsTable *tview.Table) {
    fetchAndDisplayInstances("", svc, list, detailsTable)
}

// Helper function to display the instance details.
func displayInstanceDetails(table *tview.Table, instance *ec2.Instance) {
	table.Clear()

	// Headers
	headers := []string{
		"Field", "Value",
	}

	// Assign headers to table.
	for col, header := range headers {
		cell := tview.NewTableCell(header).
			SetTextColor(tview.Styles.SecondaryTextColor).
			SetAlign(tview.AlignCenter).
			SetAttributes(tcell.AttrBold)
		table.SetCell(0, col, cell)
	}

	// Ordered list of fields
	fields := []string{
		"State",
		"Instance ID",
		"Type",
		"AMI ID",
		"Architecture",
		"Public DNS Name",
		"Public IP Address",
		"PrivateDnsName",
		"PrivateIpAddress",
		"Pem Key Name",
		"VPC ID",
		"Subnet ID",
	}

	// Mapping of fields to their pointers
	fieldMap := map[string]**string{
		"State":             &instance.State.Name,
		"Instance ID":       &instance.InstanceId,
		"Type":              &instance.InstanceType,
		"AMI ID":            &instance.ImageId,
		"Architecture":      &instance.Architecture,
		"Public DNS Name":   &instance.PublicDnsName,
		"Public IP Address": &instance.PublicIpAddress,
		"PrivateDnsName":    &instance.PrivateDnsName,
		"PrivateIpAddress":  &instance.PrivateIpAddress,
		"Pem Key Name":      &instance.KeyName,
		"VPC ID":            &instance.VpcId,
		"Subnet ID":         &instance.SubnetId,
	}

	// Add instance details to table.
	row := 1
	for _, field := range fields {
		ptr := fieldMap[field]
		value := "Unknown"
		if *ptr != nil {
			value = **ptr
		}
		table.SetCell(row, 0, tview.NewTableCell(field))
		table.SetCell(row, 1, tview.NewTableCell(value))
		row++
	}

	// Special cases that return non-string values.
	table.SetCell(row, 0, tview.NewTableCell("Iam Instance Profile Arn"))
	if instance.IamInstanceProfile != nil && instance.IamInstanceProfile.Arn != nil {
		table.SetCell(row, 1, tview.NewTableCell(*instance.IamInstanceProfile.Arn))
	} else {
		table.SetCell(row, 1, tview.NewTableCell("No ARN"))
	}
	row++

	table.SetCell(row, 0, tview.NewTableCell("Security Groups"))
	if instance.SecurityGroups != nil {
		var groupNames []string
		for _, group := range instance.SecurityGroups {
			if group != nil && group.GroupName != nil {
				groupNames = append(groupNames, *group.GroupName)
			}
		}
		table.SetCell(row, 1, tview.NewTableCell(strings.Join(groupNames, ", ")))
	} else {
		table.SetCell(row, 1, tview.NewTableCell("No Security Groups"))
	}
}

// Helper function to start an instance.
func startInstance(svc *ec2.EC2, instance *ec2.Instance) error {
	input := &ec2.StartInstancesInput{
		InstanceIds: []*string{
			instance.InstanceId,
		},
	}
	_, err := svc.StartInstances(input)
	return err
}

// Helper function to stop an instance.
func stopInstance(svc *ec2.EC2, instance *ec2.Instance) error {
	input := &ec2.StopInstancesInput{
		InstanceIds: []*string{
			instance.InstanceId,
		},
	}
	_, err := svc.StopInstances(input)
	return err
}

// Helper function to reboot an instance.
func rebootInstance(svc *ec2.EC2, instance *ec2.Instance) error {
	input := &ec2.RebootInstancesInput{
		InstanceIds: []*string{
			instance.InstanceId,
		},
	}
	_, err := svc.RebootInstances(input)
	return err
}

// Helper function to create a confirm prompt before action is taken.
func confirmModal(title, text string, confirmFunc func()) {
	modal := tview.NewModal().
		SetText(text).
		AddButtons([]string{"Yes", "No"}).
		SetDoneFunc(func(buttonIndex int, buttonLabel string) {
			if buttonLabel == "Yes" {
				confirmFunc()
			}
			app.SetRoot(flex, true)
		})
	fmt.Println("Attempting to confirm start instance")
	app.SetRoot(modal, true)
}

// Helper function to display any AWS operation errors to the user.
func handleEC2OperationResult(err error, successMsg string, list *tview.List, detailsTable *tview.Table) {
	if err != nil {
		// Show error modal
		errorModal := tview.NewModal().
			SetText("Error: " + err.Error()).
			AddButtons([]string{"Close"}).
			SetDoneFunc(func(buttonIndex int, buttonLabel string) {
				app.SetRoot(flex, true)
			})
		app.SetRoot(errorModal, true)
	} else {
		// Refresh the current instance details
		displayInstanceDetails(detailsTable, currentInstances[list.GetCurrentItem()])
	}
}

func main() {
	// Allow for profile and region from the user. Defaults provided.
	// Custom usage function for the flag package. Create a CLI help flag.
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage of %s:\n", os.Args[0])
		fmt.Println(`
	This script is an AWS EC2 Instance Viewer and Manager.
	View instance details, as well as stop, start, and reboot instances.

	Available options are:

	`)
		flag.PrintDefaults()
	}

	// AWS Profile and Region flags. Defaults provided.
	flag.StringVar(&profile, "profile", "default", "AWS Profile")
	flag.StringVar(&profile, "p", "default", "AWS Profile (shorthand)")
	flag.StringVar(&region, "region", "us-east-1", "AWS Region")
	flag.StringVar(&region, "r", "us-east-1", "AWS Region (shorthand)")
	flag.Parse()

	// If "help" flag passed, provide usage information above and exit.
	if flag.Lookup("help") != nil && flag.Lookup("help").Value.String() == "true" {
		os.Exit(0)
	}

	// Establish AWS connection.
	sess, err := session.NewSessionWithOptions(session.Options{
		Profile: profile,
		Config: aws.Config{
			Region: aws.String(region),
		},
	})
	if err != nil {
		log.Fatalf("Failed to initialize AWS session: %s", err)
	}

	svc := ec2.New(sess)

	// Create the application.
	app = tview.NewApplication()
	list := tview.NewList().ShowSecondaryText(false)
	detailsTable := tview.NewTable().SetBorders(true)
	list.SetBorder(true).SetTitle("Instances")
	detailsTable.SetBorder(true).SetTitle("Instance Details")

	initialDisplay(svc, list, detailsTable)

	// Changed function to keep index and currentInstances slice in sync during user navigation.
	list.SetChangedFunc(func(index int, mainText string, secondaryText string, shortcut rune) {
		if index >= 0 && index < len(currentInstances) {
			displayInstanceDetails(detailsTable, currentInstances[index]) 
		}
	})

	// Global view and add instance information.
	flex = tview.NewFlex().
	AddItem(list, 0, 1, true).
	AddItem(detailsTable, 0, 2, false)

	// Create the help modal.
	helpModal := tview.NewModal().
	SetText("Keyboard Commands:\n\nh - Show this help menu\nq - Quit the program\nr - Display running instances\na - Display all instances\ns - Display stopped instances\nt - Display terminated instances\n'S'(Shift+S) - Start instance.\n'X'(Shift+X) - Stop instance.\n'R'(Shift+r) - Reboot instance.").
		AddButtons([]string{"Close"}).
		SetDoneFunc(func(buttonIndex int, buttonLabel string) {
			if buttonLabel == "Close" {
				app.SetRoot(flex, true)
			}
		})

	// Assign keybindings.
	app.SetRoot(flex, true).
		SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
			switch event.Key() {
			case tcell.KeyRune:
				switch event.Rune() {
				case 'h': // Display help modal.
					app.SetRoot(helpModal, true)
					return nil
				case 'q': // Quit the application.
					app.Stop()
				case 'r': // Display running instances
					fetchAndDisplayInstances("running", svc, list, detailsTable)
					return nil
				case 'a': // Display any(all) instances.
					fetchAndDisplayInstances("any", svc, list, detailsTable)
					return nil
				case 's': // Display terminated instances.
					fetchAndDisplayInstances("stopped", svc, list, detailsTable)
					return nil
				case 't': // Display terminated instances.
					fetchAndDisplayInstances("terminated", svc, list, detailsTable)
					return nil
				case 'S': // Start Instance.
					confirmModal("Start Instance", "Are you sure you want to start the instance?", func() {
						fmt.Println("Executing start instance command")
						err := startInstance(svc, currentInstances[list.GetCurrentItem()])
						handleEC2OperationResult(err, "Successfully started the instance.", list, detailsTable)
					})
					return nil
				case 'X': // Stop Instance.
					confirmModal("Stop Instance", "Are you sure you want to stop the instance?", func() {
						err := stopInstance(svc, currentInstances[list.GetCurrentItem()])
						handleEC2OperationResult(err, "Successfully stopped the instance.", list, detailsTable)
					})
					return nil
				case 'R': // Reboot Instance.
					confirmModal("Reboot Instance", "Are you sure you want to reboot the instance?", func() {
						err := rebootInstance(svc, currentInstances[list.GetCurrentItem()])
						handleEC2OperationResult(err, "Successfully rebooted the instance.", list, detailsTable)
					})
				}
			}
			return event
		})
	
	// Exit the application on panic.
	if err := app.Run(); err != nil {
		panic(err)
	}
}
