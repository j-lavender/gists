import boto3
from datetime import datetime

# Replace with your own values
region = 'us-east-1'
date_cutoff = '2022-01-01'

# Initialize the EC2 client
ec2 = boto3.client('ec2', region_name=region)

# Get a list of all the AMIs in the account
all_amis = ec2.describe_images(Owners=['self'])

# Loop through the list and print the names and IDs of any AMIs that were created before the cutoff date
amis_to_delete = []
for ami in all_amis['Images']:
    created_date = datetime.strptime(ami['CreationDate'][:10], '%Y-%m-%d')
    cutoff_date = datetime.strptime(date_cutoff, '%Y-%m-%d')
    if created_date < cutoff_date:
        print(f"AMI {ami['ImageId']} ({ami['Name']}) was created before the cutoff date.")
        amis_to_delete.append(ami)

# Ask the user for verification before deleting the AMIs
if not amis_to_delete:
    print("No AMIs to delete.")
else:
    confirmation = input(f"\nAre you sure you want to delete {len(amis_to_delete)} AMIs? (y/n): ")
    if confirmation.lower() == 'y':
        delete_snapshots = input("Do you also want to delete the associated snapshots? (y/n): ")
        # Loop through the list of AMIs to delete and delete the associated snapshots if specified
        for ami in amis_to_delete:
            print(f"Deleting AMI {ami['ImageId']} ({ami['Name']})")
            snapshot_ids = [block_device['Ebs']['SnapshotId'] for block_device in ami['BlockDeviceMappings']]
            ec2.deregister_image(ImageId=ami['ImageId'])
            if delete_snapshots.lower() == 'y':
                for snapshot_id in snapshot_ids:
                    print(f"Deleting snapshot {snapshot_id}")
                    ec2.delete_snapshot(SnapshotId=snapshot_id)
            else:
                print(f"Skipping deletion of {len(snapshot_ids)} associated snapshots.")
    else:
        print("Aborting deletion of AMIs.")
