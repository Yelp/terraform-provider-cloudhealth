# -*- coding: utf-8 -*-
"""Dumb script to generate config from state. Not to be used without close supervision."""
from __future__ import print_function

import json
import sys

def main():
    f = open(sys.argv[1]) if len(sys.argv) > 1 and sys.argv[1] != '-' else sys.stdin

    state = json.load(f)
    for tf_name, state_resource in state['modules'][0]['resources'].items():
        type, tf_name = tf_name.split('.')
        if type != 'cloudhealth_perspective':
            continue
        print('resource "cloudhealth_perspective" "%s" {' % tf_name)

        # persp_id = state_resource['primary']['id']
        state_attr = state_resource['primary']['attributes']
        print('    name = "%s"' % state_attr['name'])
        print('    include_in_reports = %s' % state_attr['include_in_reports'])

        for group_idx in range(int(state_attr['group.#'])):
            prefix = 'group.%d.' % group_idx
            print_group(state_attr, prefix)

        print('}')


def print_group(state_attr, prefix):
    print()
    print('    group {')
    print('        name = "%s"' % state_attr[prefix + 'name'])
    print('        type = "%s"' % state_attr[prefix + 'type'])

    for rule_idx in range(int(state_attr[prefix + 'rule.#'])):
        print_rule(state_attr, prefix + 'rule.%d.' % rule_idx)

    print('    }')


def print_rule(state_attr, prefix):
    print()
    print('        rule {')
    print('            asset = "%s"' % state_attr[prefix + 'asset'])

    if state_attr[prefix + "combine_with"]:
        print('            combine_with = "%s"' % state_attr[prefix + 'combine_with'])

    print_str_list('            ', state_attr, prefix, 'field')
    print_str_list('            ', state_attr, prefix, 'tag_field')

    for cond_idx in range(int(state_attr[prefix + 'condition.#'])):
        print_condition(state_attr, prefix + 'condition.%d.' % cond_idx)

    print('        }')


def print_condition(state_attr, prefix):
    print('            condition {')
    print_str_list('                ', state_attr, prefix, 'field')
    print_str_list('                ', state_attr, prefix, 'tag_field')
    if state_attr[prefix + "op"] != '=': # is default
        print('                op = "%s"' % state_attr[prefix + "op"])
    if state_attr[prefix + "val"] != "":
        print('                val = "%s"' % state_attr[prefix + "val"])
    print('            }')


def print_str_list(indent, state_attr, prefix, field):
    count = int(state_attr[prefix + field + '.#'])
    if count == 0:
        return
    vals = '", "'.join((state_attr[prefix + field + '.' + str(x)] for x in range(count)))
    print(indent + field + ' = ["' + vals + '"]')


if __name__ == '__main__':
    main()
