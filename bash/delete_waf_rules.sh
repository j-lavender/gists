#!/bin/bash

list_waf_rules() {
    aws wafv2 list-rules --scope REGIONAL --region us-west-2
}

delete_waf_rules() {
    local rule_ids=($(aws wafv2 list-rules --scope REGIONAL --region us-west-2 --query 'Rules[].RuleId' --output text))
    
    for rule_id in "${rule_ids[@]}"; do
        echo "Deleting Rule ID: $rule_id"
        aws wafv2 delete-rule --name "$rule_id" --scope REGIONAL --region us-west-2
    done
}

list_waf_rules

read -p "Do you want to delete all the listed WAF rules? (yes/no): " response
if [[ "$response" == "yes" ]]; then
    delete_waf_rules
    echo "All WAF rules have been deleted."
else
    echo "No WAF rules have been deleted."
fi
