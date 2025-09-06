#!/usr/bin/env python3
import yaml
import sys

def add_ids_to_document(doc):
    """Add sequential IDs to rollouts and constraints in a YAML document"""
    rollout_id = 1
    constraint_id = 1
    
    # Add IDs to rollouts in flags
    if 'flags' in doc:
        for flag in doc['flags']:
            if 'rollouts' in flag and flag['rollouts']:
                for rollout in flag['rollouts']:
                    if 'id' not in rollout:
                        rollout['id'] = str(rollout_id)
                        rollout_id += 1
    
    # Add IDs to constraints in segments
    if 'segments' in doc:
        for segment in doc['segments']:
            if 'constraints' in segment and segment['constraints']:
                for constraint in segment['constraints']:
                    if 'id' not in constraint:
                        constraint['id'] = str(constraint_id)
                        constraint_id += 1
    
    return doc

def process_file(filepath):
    """Process a YAML file to add IDs"""
    with open(filepath, 'r') as f:
        doc = yaml.safe_load(f)
    
    doc = add_ids_to_document(doc)
    
    with open(filepath, 'w') as f:
        yaml.dump(doc, f, default_flow_style=False, sort_keys=False, allow_unicode=True)
    
    print(f"Processed {filepath}")

if __name__ == "__main__":
    if len(sys.argv) < 2:
        print("Usage: python add_ids.py <yaml_file>")
        sys.exit(1)
    
    for filepath in sys.argv[1:]:
        process_file(filepath)