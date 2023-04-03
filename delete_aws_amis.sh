#!/bin/bash

# Replace with your own values
region="us-east-1"
date_cutoff="2022-01-01"

# Get a list of all the AMIs in the account
all_amis=$(aws ec2 describe-images --region $region --owners self)

# Loop through the list and print the names and IDs of any AMIs that were created before the cutoff date
amis_to_delete=""
while read -r line; do
    ami_id=$(echo $line | jq -r '.ImageId')
    name=$(echo $line | jq -r '.Name')
    creation_date=$(echo $line | jq -r '.CreationDate')
    if [[ "$creation_date" < "$date_cutoff" ]]; then
        echo "AMI $ami_id ($name) was created before the cutoff date."
        amis_to_delete="$amis_to_delete $ami_id"
    fi
done <<< "$all_amis"

# Ask the user for verification before deleting the AMIs
if [[ -z "$amis_to_delete" ]]; then
    echo "No AMIs to delete."
else
    read -p "Are you sure you want to delete $(echo $amis_to_delete | wc -w) AMIs? (y/n): " confirmation
    if [[ "$confirmation" == "y" ]]; then
        read -p "Do you also want to delete the associated snapshots? (y/n): " delete_snapshots
        # Loop through the list of AMIs to delete and delete the associated snapshots if specified
        for ami_id in $amis_to_delete; do
            echo "Deleting AMI $ami_id"
            snapshot_ids=$(aws ec2 describe-images --image-ids $ami_id --region $region | jq -r '.Images[0].BlockDeviceMappings[].Ebs.SnapshotId')
            aws ec2 deregister-image --image-id $ami_id --region $region
            if [[ "$delete_snapshots" == "y" ]]; then
                for snapshot_id in $snapshot_ids; do
                    echo "Deleting snapshot $snapshot_id"
                    aws ec2 delete-snapshot --snapshot-id $snapshot_id --region $region
                done
            else
                echo "Skipping deletion of $(echo $snapshot_ids | wc -w) associated snapshots."
            fi
        done
    else
        echo "Aborting deletion of AMIs."
    fi
fi
