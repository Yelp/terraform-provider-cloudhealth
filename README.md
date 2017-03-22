# Terraform provider plugin for Cloudhealth

Supports Cloudhealth perspectives as a Terraform resource

## Install
```
$ go get github.com/bchess/terraform-provider-cloudhealth
$ make
```

Then update `~/.terraformrc`:

```
providers {
  cloudhealth = "/path/to/terraform-provider-cloudhealth/terraform-provider-cloudhealth"
}
```

## Provider Configuration
The plugin requires a Cloudhealth API key. There are two ways you can set it:

Via the environment:

```
export CHT_API_KEY=<api_key>
```

Or via ``provider.tf``:

```
provider "cloudhealth" {
    key = "<api_key>"
}
```

## Simple Perspective Example
The below example defines two groups. The first is called "My Team" who matches
against any AwsAsset with tag `team=my_team` or `team=my_team@corp.com`. The
second is a dynamic group that categorizes based on Redshift cluster name.

```
resource "cloudhealth_perspective" "my_perspective" {
    name = "My Perspective"
    include_in_reports = false

    group {
        name = "My Team"
        type = "filter"

        rule {
            asset = "AwsAsset"
            condition {
                tag_field = ["team"]
                val = "my_team"
            }
        }

        rule {
            asset = "AwsAsset"
            condition {
                tag_field = ["team"]
                val = "my_team@corp.com"
            }
        }
    }

    group {
        name = "redshift"
        type = "categorize"

        rule {
            asset = "AwsRedshiftCluster"
            field = ["Cluster Identifier"]
        }
    }

```

## Rules and Groups
Rules have the following structure.

```
rule {
    asset = <asset>
    [field = ["field1", "field2" ...]]
    [tag_field = ["tagfield1", "tagfield2" ...]

    combine_with = <[OR]|AND>
    condition {
        [field = ["field1", "field2" ...]]
        [tag_field = ["tagfield1", "tagfield2" ...]
        op = <[=]|!=|Contains|Does Not Contain,...]
        val = <val>
    }
}
```
There are only two permissible values for group type.
Static groups have `type=filter`. This is the default.
Dynamic groups have `type=categorize`. In this case you must also define `field` or `tag_field` on the rule.

### Important note about rule ordering
There is one main difference between the schema used in Terraform and the
actual Cloudhealth Perspective API.

The official API organizes a perspective by the ordering of rules, not the
order of groups. It is permissible to interleave the rules for different
groups. For example, the first rule send assets to group 1, the second rule to
send assets to group 2, then the third rule sends assets to group 1 again. This
is quite confusing and not reflected in the Perspective UI.

By comparison, the schema in Terraform groups all rules for a single group
together. All rules are ordered by the order of appearance of their groups in
the list.

If doing a `terraform import` of an existing perspective you may encounter
ordering differences in how your rules are processed. While this may require
you to make some fixes, it is my opinion that this is much more more
maintainable. It also will match the UI's presentation of the perspective
configuration.


## Not supported
Merges are not supported. Nor are dynamic groups that include additional
"filter" rules. You may get errors if you attemp to import a perspective that
has either of these things.
