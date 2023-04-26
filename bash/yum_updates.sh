#!/bin/bash

# Gather a list of AWS Instances containing their name, id, private IP address, and Key Pair name
instances=$(aws ec2 describe-instances --query 'Reservations[].Instances[].[Tags[?Key==`Name`].Value | [0], InstanceId, PrivateIpAddress, KeyName]' --output text)

# Provides that list to the user in concise, readable format
echo "Here are the instances on this AWS account:"
echo "$instances"
echo ""

# Prompt the user whether they would like to run updates on that list
read -p "Would you like to run updates on these instances? (y/n): " answer

if [[ $answer == "y" ]]; then
  # Loop over the instances and prompt the user whether to perform the updates
  for instance in $instances; do
    name=$(echo $instance | awk '{print $1}')
    id=$(echo $instance | awk '{print $2}')
    private_ip=$(echo $instance | awk '{print $3}')
    key_name=$(echo $instance | awk '{print $4}')
    
    read -p "Do you want to update $name ($private_ip)? (y/n): " update_answer
    
    if [[ $update_answer == "y" ]]; then
      # Use the private IP of the instance and Key Pair name to log in via ssh and perform the security update
      ssh -i $key_name.pem ec2-user@$private_ip "sudo yum check-updates --security"
      
      # Prompt the user to confirm before continuing
      read -p "Review the above list. Are you sure you want to update $name ($private_ip)? (y/n): " review_answer
      
      if [[ $review_answer == "y" ]]; then
        ssh -i $key_name.pem ec2-user@$private_ip "sudo yum update --security"
      fi
    fi
  done
fi

echo "Done."

# Same script but uses the public DNS record of the instance to access.
#
# Gather a list of AWS Instances containing their name, id, public IPv4 DNS, and Key Pair name
# instances=$(aws ec2 describe-instances --query 'Reservations[].Instances[].[Tags[?Key==`Name`].Value | [0], InstanceId, PublicDnsName, KeyName]' --output text)

# # Provides that list to the user in concise, readable format
# echo "Here are the instances on this AWS account:"
# echo "$instances"
# echo ""

# # Prompt the user whether they would like to run updates on that list
# read -p "Would you like to run updates on these instances? (y/n): " answer

# if [[ $answer == "y" ]]; then
#   # Loop over the instances and prompt the user whether to perform the updates
#   for instance in $instances; do
#     name=$(echo $instance | awk '{print $1}')
#     id=$(echo $instance | awk '{print $2}')
#     public_dns=$(echo $instance | awk '{print $3}')
#     key_name=$(echo $instance | awk '{print $4}')
    
#     read -p "Do you want to update $name ($public_dns)? (y/n): " update_answer
    
#     if [[ $update_answer == "y" ]]; then
#       # Use the public IPv4 DNS of the instance and Key Pair name to log in via ssh and perform the security update
#       ssh -i $key_name.pem ec2-user@$public_dns "sudo yum check-updates --security"
      
#       # Prompt the user to confirm before continuing
#       read -p "Review the above list. Are you sure you want to update $name ($public_dns)? (y/n): " review_answer
      
#       if [[ $review_answer == "y" ]]; then
#         ssh -i $key_name.pem ec2-user@$public_dns "sudo yum update --security"
#       fi
#     fi
#   done
# fi

# echo "Done."
