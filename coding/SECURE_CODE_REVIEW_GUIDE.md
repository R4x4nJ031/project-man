# Secure Code Review Guide

This guide is meant to teach a reusable way to approach secure code review, not just for `project-man`, but for backend projects in general.

The biggest mistake people make is reviewing code line-by-line without first understanding:

- what the system does
- what it trusts
- what it must protect
- where failure would matter most

Secure code review gets much easier once you review in layers.

## 1. What Secure Code Review Really Is

Secure code review is not:

- reading random files and hoping to spot something scary
- searching only for obvious bad functions
- checking syntax style

Secure code review is:

- understanding the system
- identifying trust boundaries
- finding places where untrusted input becomes trusted too early
- checking whether controls actually enforce the intended security properties

In simple terms:

You are asking, “How could this system be made to do something unsafe, and what in the code prevents that?”

## 2. The Right Review Order

This is the order I recommend for almost any backend project.

### Step 1: Understand the Product

Before reading deep code, answer:

- what does the app do?
- who uses it?
- what data does it protect?
- what actions are sensitive?

If you skip this, you will miss business logic issues.

### Step 2: Map the Architecture

Identify:

- entrypoints
- middleware
- service layer
- repositories / DB access
- external services
- config loading
- test paths

You want to know the request flow before judging individual files.

### Step 3: Find Trust Boundaries

Ask:

- what inputs come from users?
- what comes from tokens?
- what comes from headers?
- what comes from DB lookups?
- what becomes trusted only after validation?

This is where many real bugs begin.

### Step 4: Review High-Risk Paths First

Don’t start with helpers.

Start with:

- auth
- tenant/workspace context
- authorization
- file access
- payment logic
- admin actions
- sensitive mutations
- DB access patterns

This gets you to real risk faster.

### Step 5: Review Defense In Depth

Ask:

- if one layer fails, what catches it?
- does the DB protect data?
- does the app protect data?
- are dangerous actions audited?
- do tests prove the control works?

This is where mature systems stand out.

## 3. The Main Categories To Review

When reviewing code, mentally group issues into these buckets.

### Authentication

Questions:

- how is identity established?
- what validates the token/session?
- are signatures checked?
- are issuer and audience checked?
- can identity be spoofed?

### Authorization

Questions:

- who is allowed to do what?
- where is that decided?
- is the model deny-by-default?
- can a lower-privilege user reach a high-privilege action?

### Tenant / Resource Isolation

Questions:

- how does the app know which tenant or resource scope is active?
- is resource ownership verified server-side?
- can a user access another tenant’s data with guessed IDs?

### Input Validation

Questions:

- are headers/body/query params validated?
- do type and semantic checks exist?
- can malformed input alter logic flow?

### Database Safety

Questions:

- are queries scoped correctly?
- are dangerous assumptions hidden in repositories?
- are migrations weakening protections?
- is least privilege enforced?
- does RLS or another DB-layer control exist where needed?

### Auditability

Questions:

- are sensitive mutations logged?
- is actor identity captured?
- is logging transactional where it needs to be?

### Secrets / Configuration

Questions:

- are secrets hardcoded?
- are dev helpers clearly separated?
- are config defaults misleading or risky?

### Operational Safety

Questions:

- are scripts using the right setup path?
- do tests run against the intended environment?
- are controls actually exercised in CI or only documented?

## 4. File Structure Review Strategy

A fast and effective way to review an unfamiliar repo is:

1. `README`
2. threat model / docs
3. entrypoint
4. middleware
5. handlers
6. services
7. repositories
8. migrations
9. tests
10. scripts

Why this works:

- docs tell you intent
- entrypoint tells you request flow
- middleware tells you security boundaries
- handlers tell you input handling
- services tell you business rules
- repositories tell you data access safety
- migrations tell you database security reality
- tests tell you what the team actually proved

## 5. How To Think About Business Logic Bugs

This is where many people stay too shallow.

Security review is not only:

- SQL injection
- hardcoded secrets
- XSS

It is also:

- cross-tenant update
- missing ownership check
- role escalation
- unreviewed admin path
- audit gap
- inconsistent behavior between routes

A strong reviewer always asks:

- What business rule is this code supposed to preserve?
- Can the user violate it by combining requests in a weird way?

That is product security thinking.

## 6. The Questions That Find Real Bugs

Here are high-value questions that often uncover real issues:

### Identity Questions

- Can I become another user?
- Can I use an expired or malformed token?
- Can I confuse the parser or algorithm validation?

### Scope Questions

- Can I choose a tenant/resource I should not control?
- Is the server verifying that scope, or only trusting client input?

### Authorization Questions

- Can a lower role call this mutation anyway?
- Is there another route that reaches the same action with weaker checks?

### Object Access Questions

- If I know an ID, can I read or mutate it?
- Does the server verify ownership or tenant membership at the data layer?

### Consistency Questions

- Do create, update, and delete use the same controls?
- Is one route stricter than another for the same resource?

### Audit Questions

- If something sensitive changes, can we prove who did it?
- Can a write happen without an audit trail?

### DB Questions

- If a developer forgets a filter, what happens?
- Can the DB role bypass security controls?

## 7. Common Review Mistakes

Avoid these.

### Mistake 1: Reading too locally

If you stare at one function without understanding the request flow, you miss bigger issues.

### Mistake 2: Only searching for famous vulnerability names

Many serious bugs are business-logic bugs, not textbook injection bugs.

### Mistake 3: Trusting comments over behavior

Review what the code and tests actually enforce.

### Mistake 4: Confusing “works” with “secure”

A route returning the correct happy-path response tells you almost nothing about abuse resistance.

### Mistake 5: Ignoring the database role and schema

Real security often depends on DB configuration, not just handlers.

## 8. How To Write Good Findings

A useful review finding has:

- what is wrong
- why it matters
- where it is
- how it can fail

Good structure:

1. issue
2. impact
3. evidence
4. expected fix direction

Example:

- “The PostgreSQL application role has `BYPASSRLS`, which makes tenant RLS ineffective. That allows cross-tenant reads if an application query is written without explicit tenant filters.”

That is much stronger than:

- “RLS maybe broken”

## 9. Review Modes

Use different modes depending on context.

### Fast Triage Review

Use when time is short.

Focus on:

- auth
- authz
- tenant/resource isolation
- DB role and schema protections
- audit trail

### Deep Product Security Review

Use when the system is important or the feature is risky.

Focus on:

- business logic invariants
- abuse cases
- alternate routes to the same action
- operational setup
- test coverage of controls

### Regression Review

Use after a bug fix or security improvement.

Focus on:

- whether the root cause is actually closed
- whether a test proves the new behavior
- whether related paths remain vulnerable

## 10. How To Practice This Skill

If you want to get strong fast, use this loop:

1. read the README
2. write down assets and trust boundaries
3. predict likely bugs before reading code
4. read code and see if your predictions were right
5. verify with tests or runtime checks
6. write findings clearly

This turns review into a learnable process instead of a talent test.

## 11. Applying This To `project-man`

If you review this repo well, the top areas to inspect are:

- JWT validation
- `X-Tenant-ID` trust transition
- membership lookup
- RBAC checks
- project repository queries
- RLS migrations and DB role hardening
- transactional audit logging
- live tests and integration tests

That is why this project is a good practice target for secure review.

## 12. What “Mastery” Looks Like

You do not need to memorize every bug class.

Mastery looks more like this:

- you can find trust boundaries quickly
- you know where to read first
- you ask the right questions about identity, scope, and business rules
- you can verify whether controls are real or only assumed
- you can explain findings clearly and calmly

That is what strong product security engineers do in practice.

## 13. A Reusable Review Checklist In One Page

If you need the shortest version possible, use this:

1. What are we protecting?
2. Who can act on it?
3. What input is untrusted?
4. When does it become trusted?
5. What enforces identity?
6. What enforces authorization?
7. What enforces tenant/resource isolation?
8. What protects the DB if app code is wrong?
9. What is audited?
10. What remains risky?

If you can answer those 10 questions well on a codebase, you are doing real secure code review.
