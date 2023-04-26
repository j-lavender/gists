import boto3

def list_orphaned_snapshots():
    # Connect to AWS using Boto3
    ec2 = boto3.resource('ec2')

    # Get a list of all snapshots
    snapshots = ec2.snapshots.all()

    # Get a list of all volumes
    volumes = ec2.volumes.all()

    # Create a set to hold the volume IDs
    volume_ids = set()

    # Add all volume IDs to the set
    for volume in volumes:
        volume_ids.add(volume.id)

    # Create a list to hold the orphaned snapshots
    orphaned_snapshots = []

    # Check each snapshot to see if its volume ID is in the set
    for snapshot in snapshots:
        if snapshot.volume_id not in volume_ids:
            orphaned_snapshots.append(snapshot.id)

    # Return the list of orphaned snapshots
    return orphaned_snapshots

def delete_snapshots(snapshot_ids):
    # Connect to AWS using Boto3
    ec2 = boto3.resource('ec2')

    # Confirm with user before deleting snapshots
    confirmation = input(f"Are you sure you want to delete {len(snapshot_ids)} snapshots? (y/n): ")
    if confirmation.lower() != "y":
        print("Aborted snapshot deletion.")
        return

    # Delete each snapshot in the list
    for snapshot_id in snapshot_ids:
        snapshot = ec2.Snapshot(snapshot_id)
        snapshot.delete()
        print(f"Deleted snapshot {snapshot_id}")

# Call the function to list orphaned snapshots
orphaned_snapshots = list_orphaned_snapshots()

# Print the list of orphaned snapshots
print("Orphaned Snapshots:")
for snapshot_id in orphaned_snapshots:
    print(snapshot_id)

# Call the function to delete the orphaned snapshots
delete_snapshots(orphaned_snapshots)
