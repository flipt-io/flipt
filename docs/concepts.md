# Concepts

This document describes the basic concepts of Flipt. More information on how to use Flipt is noted in the [Getting Started](getting_started.md) documentation.

## Flags

Flags are the basic unit in the Flipt ecosystem. Flags represent experiments or features that you want to be able to enable or disable for users of your applications.

For example, a flag named `new-contact-page`, could be used to determine whether or not a given user sees the latest version of a contact us page that you are working on when they visit your homepage.

Flags can be used as simple on/off toggles or with variants and rules to support more elaborate usecases.

![Flags Example](assets/images/concepts/00_flags.png?raw=true "Flags Example")

## Variants

Variants are options for flags. For example, if you have a flag `colorscheme` that determines which main colors your users see when they log in to your application, then possible variants could be include `blue`, `green` or `red`.

![Variants Example](assets/images/concepts/01_variants.png?raw=true "Variant Example")

!!! note
    Variant keys must be unique for a given flag.

## Segments

Segments allow you to split your userbase or audience up into predefined slices. This is a powerful feature that enables targeting groups to determine if a flag or variant applies to them.

An example segment could be `new-users`.

![New Users Segment](assets/images/concepts/02_segments.png)

!!! tip
    Segments are global across the Flipt application so they can be used with multiple flags.

## Constraints

Constraints allow you to determine which segment a given entity is a part of.

For example, for a user to fall into the above `new-users` segment, you may want to check their `finished_onboarding` property.

![Constraints Example](assets/images/concepts/03_constraints.png?raw=true "Constraints Example")

All constraints have a *property*, *type*, *operator* and optionally a *value*.

<dl>
<dt><strong>property</strong></dt>
<dd>the context key to match against, see the context section below</dd>
<dt><strong>type</strong></dt>
<dd>one of the basic types: string, number or boolean</dd>
<dt><strong>operator</strong></dt>
<dd>how to compare the property against the value</dd>
<dt><strong>value</strong> (optional)</dt>
<dd>what to compare with the operator<dd>
</dl>

!!! note
    In order for a segment to match, it must match **ALL** of it's constraints.

## Rules

Rules allow you to tie your flags, variants and segments together by specifying which segments are targeted by which variants.

Rules can be as simple as `IF IN segment THEN RETURN variant_a` or they can be more rich by using distribution logic to rollout features on a percent basis.

Continuing our previous example, we may want to return the flag variant `blue` for all entities in the `new-users` segment. This would be configured like so:

![Rules Example](assets/images/concepts/04_rules.png?raw=true "Rules Example")

!!! note
    As shown, rules are evaluated in order per their rank from 1-N. The first rule that matches wins. Once created, rules can be re-ordered to change how they are evaluated.

## Distributions

Distributions allow you to rollout different variants of your flag to percentages of your userbase based on your rules.

Let's say that instead of always showing the `blue` variant to your `new-users` segment, you want to show blue to 30% of `new-users`, `red` to 10%, and `green` to the remaining 60%. You would accomplish this using rules with distributions:

![Distributions Example](assets/images/concepts/05_distributions.png?raw=true "Distributions Example")

This is an extremely powerful feature of Flipt that can help you seamlessly deploy new features of your applications to your users while also limiting reach of potential bugs.

## Evaluation

Evaluation is the process of sending requests to the Flipt server to process and determine if that request matches any of your segments, and if so which variant to return.

In the above example involving colors, evaluation is where you send information about your current user to determine if they are a `new-user`, and which color (`blue`, `red`, or `green`) that they should see for their main colorscheme.

### Entities

Evaluation works by uniquely identifying each _thing_ that you want to compare against your segments and flags. We call this an `entity` in the Flipt ecosystem. More often than not this will be a user, but we didn't want to make any assumptions on how your application works, which is why `entity` was chosen.

<dl>
<dt><strong>entity</strong></dt>
<dd>what you want to test against in your application</dd>
</dl>

For Flipt to successfully determine which _bucket_ your entities fall into, it must have a way to uniquely identify them. This is the `entityId` and it is a simple string. It's up to you what that `entityId` is.

It could be a:

* email address
* userID
* ip address
* physical address
* etc

Anything that is unique enough for your application and it's requirements.

### Context

The final piece of the puzzle is context. Context allows Flipt to determine which segment your entity falls into by comparing it to all of the possible constraints that you defined.

<dl>
<dt><strong>context</strong></dt>
<dd>metadata associated with your entity, used to determine which if any segments that entity is a member of</dd>
</dl>

Examples of context could include:

* isAdmin
* favoriteColor
* country
* freeUser

Think of these as pieces of information that are usually not unique, but that can be used to split your entities into your segments.

You can include as much or as little context for each entity as you want, however the more context that you provide, the more likely it is that a entity will match one of your segments.

In Flipt, `context` is a simple map of key-value pairs where the key is the property to match against all constraints, and the value is what is compared.
