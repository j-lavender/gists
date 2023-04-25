import boto3

# Source AWS profile and security group information
source_profile_name = "SOURCE_PROFILE_NAME"
source_region_name = "SOURCE_REGION_NAME"
source_security_group_id = "SOURCE_SECURITY_GROUP_ID"

# Destination AWS profile and VPC information
destination_profile_name = "DESTINATION_PROFILE_NAME"
destination_region_name = "DESTINATION_REGION_NAME"
destination_vpc_id = "DESTINATION_VPC_ID"

# Create boto3 sessions for source and destination AWS profiles
source_session = boto3.Session(profile_name=source_profile_name, region_name=source_region_name)
destination_session = boto3.Session(profile_name=destination_profile_name, region_name=destination_region_name)

# Create boto3 clients for source and destination AWS accounts
source_ec2 = source_session.client('ec2')
destination_ec2 = destination_session.client('ec2')

# Retrieve source security group information
response = source_ec2.describe_security_groups(GroupIds=[source_security_group_id])

# Create a new security group in the destination VPC with the same name as the source security group
new_security_group = destination_ec2.create_security_group(GroupName=response['SecurityGroups'][0]['GroupName'], 
                                                            Description=response['SecurityGroups'][0]['Description'], 
                                                            VpcId=destination_vpc_id)

# Add the inbound and outbound rules from the source security group to the new security group
for ip_permission in response['SecurityGroups'][0]['IpPermissions']:
    try:
        destination_ec2.authorize_security_group_ingress(GroupId=new_security_group['GroupId'], 
                                                         IpPermissions=[ip_permission])
    except destination_ec2.exceptions.ClientError as e:
        if e.response['Error']['Code'] == 'InvalidPermission.Duplicate':
            print(f"Rule already exists in new security group: {ip_permission}")
        else:
            raise e

for ip_permission in response['SecurityGroups'][0]['IpPermissionsEgress']:
    try:
        destination_ec2.authorize_security_group_egress(GroupId=new_security_group['GroupId'], 
                                                        IpPermissions=[ip_permission])
    except destination_ec2.exceptions.ClientError as e:
        if e.response['Error']['Code'] == 'InvalidPermission.Duplicate':
            print(f"Rule already exists in new security group: {ip_permission}")
        else:
            raise e

print(f"Security group copied successfully! New security group ID: {new_security_group['GroupId']}")
