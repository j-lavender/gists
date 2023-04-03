#!/bin/bash

# List orphaned snapshots in AWS

# Set AWS profile and region
export AWS_PROFILE=my_profile
export AWS_DEFAULT_REGION=us-east-1

# Get a list of all snapshots
snapshots=$(aws ec2 describe-snapshots --owner-ids self --query 'Snapshots[*].[SnapshotId,VolumeId]' --output text)

# Get a list of all volumes
volumes=$(aws ec2 describe-volumes --query 'Volumes[*].VolumeId' --output text)

# Create an array to hold the volume IDs
volume_ids=()

# Add all volume IDs to the array
for volume in $volumes; do
  volume_ids+=($volume)
done

# Create an array to hold the orphaned snapshots
orphaned_snapshots=()

# Check each snapshot to see if its volume ID is in the array
while read snapshot; do
  snapshot_id=$(echo $snapshot | awk '{print $1}')
  volume_id=$(echo $snapshot | awk '{print $2}')
  if [[ ! " ${volume_ids[@]} " =~ " ${volume_id} " ]]; then
    orphaned_snapshots+=($snapshot_id)
  fi
done <<< "$snapshots"

# Print the list of orphaned snapshots
echo "Orphaned Snapshots:"
for snapshot_id in "${orphaned_snapshots[@]}"; do
  echo $snapshot_id
done

# Delete the orphaned snapshots
read -p "Are you sure you want to delete ${#orphaned_snapshots[@]} snapshots? (y/n): " confirmation
if [[ $confirmation =~ ^[Yy]$ ]]; then
  for snapshot_id in "${orphaned_snapshots[@]}"; do
    aws ec2 delete-snapshot --snapshot-id $snapshot_id
    echo "Deleted snapshot $snapshot_id"
  done
else
  echo "Aborted snapshot deletion."
fi
