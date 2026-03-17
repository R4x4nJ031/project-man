# Threat Model

This document describes the threat model for `project-man`.

The goal is not to make the project look scary. The goal is to make security reasoning explicit:

- what we are protecting
- who can interact with the system
- where trust boundaries exist
- what can go wrong
- what controls exist today
- what residual risks remain

This is the same mindset used in product security work. A threat model is not just a document. It is a way of thinking about a system before and during implementation.

## 1. System Overview

`project-man` is a multi-tenant backend for managing projects.

Main capabilities:

- authenticate users with JWT
- resolve the active tenant through membership checks
- enforce RBAC for project actions
- store projects in PostgreSQL
- enforce tenant isolation in both application code and PostgreSQL RLS
- audit project mutations

Primary components:

- HTTP API server
- JWT auth middleware
- tenant membership middleware
- RBAC middleware
- PostgreSQL database
- Goose migrations
- audit logging layer

## 2. Security Objectives

These are the core properties the system is trying to preserve.

### Identity Integrity

Only requests with valid JWTs should reach protected business logic.

### Tenant Isolation

A user in one tenant must not be able to read or modify another tenant's project data.

### Authorization Integrity

Users should only perform actions allowed by their role in the active tenant.

### Mutation Traceability

Every sensitive project mutation should be attributable to an actor and tenant.

### Database Defense in Depth

Even if application code forgets tenant filtering, PostgreSQL should still restrict access through RLS.

## 3. Assets

Assets are the things worth protecting.

### Project Data

- project IDs
- project names
- project creation timestamps
- tenant association

### Tenant Membership Data

- which users belong to which tenants
- what role each user has in each tenant

### Audit Logs

- actor identity
- action performed
- resource touched
- tenant scope
- mutation time

### Authentication Secrets

- JWT signing secret
- JWT issuer and audience configuration

### Database Integrity

- schema state
- RLS policies
- role hardening

## 4. Actors

Actors are the entities that interact with the system.

### Legitimate User

A normal user with a valid JWT and one or more tenant memberships.

### Malicious Authenticated User

A valid user attempting:

- cross-tenant access
- role abuse
- IDOR-style resource access
- unauthorized mutations

### Unauthenticated External Attacker

A party without a valid JWT attempting:

- protected route access
- malformed request abuse
- token forgery

### Developer / Maintainer

A trusted internal actor who can:

- add code
- change queries
- alter migrations

This actor matters in the threat model because mistakes by trusted developers are a major source of security bugs.

### Database Role

The PostgreSQL application role itself is a security actor in practice because its privileges determine whether RLS is meaningful.

## 5. Entry Points

These are the inputs where threats can enter.

### HTTP Endpoints

- `GET /healthz`
- `GET /secure`
- `POST /projects`
- `GET /projects/list`
- `GET /projects/get`
- `PUT /projects/update`
- `DELETE /projects/delete`

### Request Headers

- `Authorization`
- `X-Tenant-ID`

### Request Bodies

- JSON payloads for create/update

### Environment Variables

- `DATABASE_URL`
- `JWT_SECRET`
- `JWT_ISSUER`
- `JWT_AUDIENCE`
- `PORT`

### Database Layer

- application queries
- transaction-local tenant settings
- Goose migrations

## 6. Trust Boundaries

Trust boundaries are where data crosses from less trusted space into more trusted space.

This is one of the most important parts of threat modeling.

### Boundary 1: Client to API Server

The client controls:

- headers
- request bodies
- query parameters

These values must be validated and never assumed to be trustworthy.

### Boundary 2: JWT to Application Identity

The server treats identity claims as trusted only after:

- signature validation
- expiry validation
- issuer validation
- audience validation

### Boundary 3: Tenant Header to Tenant Context

`X-Tenant-ID` is client-controlled input.

It becomes trusted tenant context only after the server verifies membership in the database.

### Boundary 4: Application to Database

The application sends queries to PostgreSQL.

This is a trust boundary because:

- missing filters can expose data
- DB role misconfiguration can weaken protections
- transaction handling determines whether tenant settings are safely scoped

### Boundary 5: Business Mutation to Audit Trail

Audit data is trustworthy only if it is written in the same transaction as the business mutation.

Otherwise the system could claim an action happened without evidence, or write evidence for a failed action.

## 7. Main Threats

This section focuses on the highest-value threats for this project.

### Threat 1: Unauthenticated Access to Protected Routes

Example:

- attacker sends requests without a JWT
- attacker sends forged or malformed tokens

Impact:

- unauthorized access to project operations

Controls:

- JWT Bearer auth middleware
- signing method validation
- expiry check
- issuer check
- audience check

Residual risk:

- if `JWT_SECRET` is weak or leaked, token forgery becomes possible

### Threat 2: Cross-Tenant Read Access

Example:

- user from `tenant-acme` tries to read data from `tenant-beta`
- developer forgets a tenant filter in a query

Impact:

- direct tenant data exposure

Controls:

- membership check in tenant middleware
- tenant-scoped application queries
- PostgreSQL RLS policy on `projects`
- PostgreSQL role hardened with `NOBYPASSRLS`
- integration test proving direct cross-tenant read is blocked

Residual risk:

- other future tables without RLS would still need protection

### Threat 3: Cross-Tenant Mutation

Example:

- user attempts update/delete using another tenant's project ID

Impact:

- unauthorized modification or deletion of data

Controls:

- membership verification
- RBAC before mutation
- tenant-scoped queries
- RLS `USING` and `WITH CHECK`
- current behavior returns `404` to reduce existence leakage

Residual risk:

- future mutation paths must also use tenant-scoped transactions correctly

### Threat 4: Role Abuse / Privilege Escalation

Example:

- `viewer` attempts to create or update projects
- user tampers with client behavior hoping to bypass role restrictions

Impact:

- unauthorized state changes

Controls:

- role resolved server-side from memberships
- permission middleware
- deny-by-default behavior for unknown roles

Residual risk:

- if role mapping changes incorrectly in code, authorization could weaken

### Threat 5: IDOR (Insecure Direct Object Reference)

Example:

- attacker guesses a project ID from another tenant
- sends `GET /projects/get?id=<other-tenant-id>`

Impact:

- data leakage or unauthorized mutation

Controls:

- tenant-scoped queries
- RLS on `projects`
- cross-tenant misses return `404`

Residual risk:

- any future resource type must follow the same pattern

### Threat 6: Audit Trail Gaps

Example:

- mutation succeeds but audit write fails
- audit row is written even though mutation rolls back

Impact:

- unreliable forensics
- weak incident response

Controls:

- audit writes occur in the same transaction as the mutation
- create/update/delete services wrap business mutation and audit insert together

Residual risk:

- read access is not audited by design

### Threat 7: RLS Misconfiguration

Example:

- RLS policy exists but DB role can bypass it
- session tenant context leaks across pooled connections

Impact:

- false sense of security
- real tenant isolation failure

Controls:

- `ALTER ROLE project NOSUPERUSER NOBYPASSRLS`
- `FORCE ROW LEVEL SECURITY`
- tenant context set with transaction-local `set_config(..., true)`
- integration test proving RLS enforcement

Residual risk:

- future DB roles must be reviewed with the same care

### Threat 8: Sensitive Configuration Exposure

Example:

- hardcoded secrets
- logging of secrets
- accidental secret commits

Impact:

- auth compromise
- environment compromise

Controls:

- env-based config loading
- no normal-path secret logging

Residual risk:

- local development helper token generation still assumes a known test secret
- CI secret handling is not yet implemented

## 8. Abuse Cases

Abuse cases are concrete attacker behaviors we want to think through and test.

### Abuse Case A: Missing Auth Header

Expected result:

- `401`
- request blocked before business logic

### Abuse Case B: Invalid JWT

Expected result:

- `401`

### Abuse Case C: Missing Tenant Header

Expected result:

- `403`

### Abuse Case D: Unknown Tenant Membership

Expected result:

- `403`

### Abuse Case E: Viewer Attempts Create

Expected result:

- `403`

### Abuse Case F: Cross-Tenant Update by Resource ID

Expected result:

- `404`

### Abuse Case G: Direct Cross-Tenant DB Read

Expected result:

- no row returned due to RLS

This is especially important because it proves the database layer is enforcing tenant isolation, not just the handlers.

## 9. Existing Controls Summary

Current controls already implemented:

- JWT authentication
- issuer and audience validation
- membership-based tenant resolution
- RBAC middleware
- tenant-scoped DB transactions
- PostgreSQL RLS
- hardened DB role with `NOBYPASSRLS`
- audit logging for mutations
- Goose migrations
- live endpoint abuse-case testing
- unit tests
- integration tests for RLS and audit logging

## 10. Residual Risks

Threat modeling is not just about controls. It is also about being honest about what still remains.

Current residual risks:

- tenant selection is still client-provided through `X-Tenant-ID`, then server-validated
- only `projects` currently has RLS protection
- no CI enforcement yet
- no file storage or object access controls yet
- no threat model for future features like file upload
- no rate limiting
- no request ID propagation yet
- no structured security event logging beyond audit records

These are not failures. They are simply known next-step areas.

## 11. What This Teaches

This project is a good threat modeling exercise because it shows the difference between:

- business logic checks
- authorization checks
- tenant isolation checks
- database-enforced protections

The biggest lesson is:

Do not only ask:

- "Does my code work?"

Also ask:

- "If a developer makes a mistake later, what still protects the system?"

That is the core product security mindset.

## 12. Practical Threat Modeling Method

For future projects, this is a simple repeatable method:

1. Define the system in one paragraph.
2. List the assets.
3. List the actors.
4. Mark the trust boundaries.
5. Ask what each actor could do at each entry point.
6. Focus on confidentiality, integrity, and abuse potential.
7. Write down existing controls.
8. Write down what still remains risky.

If you can do those 8 steps clearly, you are already doing useful threat modeling.
