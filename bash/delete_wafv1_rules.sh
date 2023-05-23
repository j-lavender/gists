#!/bin/bash

get_change_token() {
    local response=$(aws waf get-change-token --region us-west-2)
    local change_token=$(echo "$response" | jq -r '.ChangeToken')
    echo "$change_token"
}

list_waf_rules() {
    local next_token=""
    
    while true; do
        local response=$(aws waf list-rules --region us-west-2 --limit 50 --next-marker "$next_token")
        local rule_ids=($(echo "$response" | jq -r '.Rules[].RuleId'))
        
        for rule_id in "${rule_ids[@]}"; do
            local rule_name=$(aws waf get-rule --region us-west-2 --rule-id "$rule_id" --query 'Rule.Name')
            echo "Rule ID: $rule_id, Rule Name: $rule_name"
        done
        
        local next_token=$(echo "$response" | jq -r '.NextMarker')
        if [[ $next_token == "null" ]]; then
            break
        fi
    done
}

delete_waf_rules() {
    local change_token=$(get_change_token)
    local next_token=""
    
    while true; do
        local response=$(aws waf list-rules --region us-west-2 --limit 50 --next-marker "$next_token")
        local rule_ids=($(echo "$response" | jq -r '.Rules[].RuleId'))
        
        for rule_id in "${rule_ids[@]}"; do
            echo "Deleting Rule ID: $rule_id"
            aws waf delete-rule --region us-west-2 --rule-id "$rule_id" --change-token "$change_token"
        done
        
        local next_token=$(echo "$response" | jq -r '.NextMarker')
        if [[ $next_token == "null" ]]; then
            break
        fi
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
