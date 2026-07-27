# Correctness Properties — Worked Examples

Worked examples for each Correctness Property category. Referenced from `design.md` §2.6.

Each example shows (1) a requirement, (2) the §2.6 property definition, (3) the §2.8 property-based test table entry, and (4) what the generator does.

---

## Equivalence

Suppose REQ-1.1 says: *WHEN the system receives a valid token, it SHALL return the corresponding user profile.*

**§2.6 definition:**

```
Property 1: Token-to-profile resolution
Category: Equivalence
Statement: For all valid tokens T issued for user U, resolveProfile(T) = profile(U)
Validates: Requirements 1.1
```

**§2.8 property-based test table entry:**

| Test | Property | Generator description | Tags |
|------|----------|-----------------------|------|
| `prop_token_resolves_to_owner_profile` | Property 1 | Generate random valid User, issue a token for that user, resolve profile via token | `Property/1` |

The generator creates a random user, issues a token, and asserts that resolving the token returns the same user's profile. This verifies the equivalence property across many random inputs.

---

## Absence

Suppose REQ-2.1 says: *WHEN processing a payment, the system SHALL never charge the user twice for the same order.*

**§2.6 definition:**

```
Property 2: No duplicate charges
Category: Absence
Statement: For all orders O processed concurrently, chargeCount(O) = 1
Validates: Requirements 2.1
```

**§2.8 property-based test table entry:**

| Test | Property | Generator description | Tags |
|------|----------|-----------------------|------|
| `prop_no_duplicate_charges` | Property 2 | Generate random Order, simulate N concurrent payment attempts for that order, count successful charges | `Property/2` |

The generator creates a random order, fires multiple concurrent payment requests, and asserts the total charge count is exactly 1. This verifies the absence of duplicate charges under concurrency.

---

## Round-trip

Suppose REQ-3.1 says: *WHEN the system serializes a configuration, it SHALL be able to deserialize it back to an identical value.*

**§2.6 definition:**

```
Property 3: Config serialization round-trip
Category: Round-trip
Statement: For all valid configurations C, deserialize(serialize(C)) = C
Validates: Requirements 3.1
```

**§2.8 property-based test table entry:**

| Test | Property | Generator description | Tags |
|------|----------|-----------------------|------|
| `prop_config_roundtrip` | Property 3 | Generate random valid Config with varied field types, serialize to storage format, deserialize back | `Property/3` |

The generator creates a random valid configuration, serializes it, deserializes the result, and asserts structural equality with the original. This verifies no data loss or mutation during the round-trip.

---

## Propagation

Suppose REQ-4.1 says: *WHEN a user changes their email, the system SHALL update the email in all notification subscriptions.*

**§2.6 definition:**

```
Property 4: Email propagation to subscriptions
Category: Propagation
Statement: For all users U and new emails E, after updateEmail(U, E), every subscription of U has email = E
Validates: Requirements 4.1
```

**§2.8 property-based test table entry:**

| Test | Property | Generator description | Tags |
|------|----------|-----------------------|------|
| `prop_email_propagates_to_subscriptions` | Property 4 | Generate random User with 1–10 subscriptions, update email to a new random value, inspect all subscriptions | `Property/4` |

The generator creates a user with multiple subscriptions, changes the user's email, and asserts every subscription reflects the new address. This verifies the change propagates to all dependent records.

---

## Exclusion

Suppose REQ-5.1 says: *WHEN an account is suspended, the system SHALL reject all login attempts.*

**§2.6 definition:**

```
Property 5: Suspended accounts cannot be active
Category: Exclusion
Statement: For all accounts A, suspended(A) and activeSession(A) are never both true
Validates: Requirements 5.1
```

**§2.8 property-based test table entry:**

| Test | Property | Generator description | Tags |
|------|----------|-----------------------|------|
| `prop_suspended_blocks_login` | Property 5 | Generate random Account, suspend it, attempt login, inspect account state and session | `Property/5` |

The generator creates a random account, suspends it, attempts a login, and asserts that login fails and no active session exists. This verifies the two states are mutually exclusive.
