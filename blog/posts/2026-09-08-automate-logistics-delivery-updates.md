---
title: Automate Delivery Updates for Online Store Orders
slug: automate-logistics-delivery-updates
description: "Automate delivery updates for online store orders: courier data quality, failed attempts, cash on delivery reattempts, and why a sent message is not a delivered one."
type: general
date: 2026-09-08
related: [where-is-my-order-whatsapp, whatsapp-ai-shopify-lebanon-gulf, how-to-automate-whatsapp-for-your-business]
ctaLine: Losing hours to delivery chasing? Ask plum how order events, courier status and the customer's own thread can run as one loop.
keywords: [automate delivery updates, courier tracking automation, logistics automation ecommerce, cash on delivery operations, order tracking WhatsApp]
hero: /blog/heroes/wrapped-parcel-plain.jpg
heroAlt: A plainly wrapped parcel tied with string
heroCredit: Rawpixel Ltd
heroCreditUrl: https://www.flickr.com/photos/147875007@N03
heroSource: Flickr
heroLicense: CC0 1.0
heroLicenseUrl: https://creativecommons.org/publicdomain/zero/1.0/
notionUrl: https://app.notion.com/p/3d58d6e734f4819d8f93fa8870b99a5a
faq:
  - q: Can I get courier tracking data automatically?
    a: Usually yes, through the carrier's API. Aramex, for example, publishes a shipments tracking API in its developer solutions centre that accepts a list of waybill numbers and returns their tracking data, requiring a registered account with credentials. Access is typically pull based rather than pushed to you, so how fresh your data is depends on how often you poll, and carriers differ in what they expose.
  - q: Does a delivered status on WhatsApp mean the parcel arrived?
    a: "No. Meta's status webhook describes your message, not your shipment. Delivered means the message reached at least one of the customer's devices. A message can stay at sent because the customer has not come online, and Meta holds messages for offline customers for a 30 day window. Track parcel state and message state as two separate things."
  - q: Should I send an out for delivery message?
    a: Only if your courier updates that status early enough on the delivery day to be useful. A notification that arrives after the parcel has already been delivered is a paid message that makes your business look disorganised. If the carrier's timing is unreliable, send the dispatch notification and skip this one.
  - q: How should automation handle a failed delivery attempt?
    a: By escalating, not by messaging alone. Map the failure status to an internal alert with a named owner and a deadline measured in hours, because whether a reattempt is worth a second leg is a commercial decision, particularly on cash on delivery. Contact the customer with a concrete next option rather than an apology, and log the reason in a fixed vocabulary so the pattern behind repeated failures becomes visible.
  - q: Do delivery notifications cost money?
    a: A notification you send first, when no customer service window is open, is a utility template and is billed at the rate for that customer's market. Since 1 October 2026, replies inside the 24 hour window a customer opened are free for the first 1,000 per business phone number each month, then charged, and utility templates are charged even when sent in response to a user. Meta revised its rate cards effective 1 July 2026 and rates are market specific, so check the current card rather than a quoted figure.
  - q: What changes when orders are cash on delivery?
    a: The delivery attempt is also the payment attempt, so a failed attempt is a cancelled sale plus two legs of shipping rather than a rescheduling. Confirm the order and the delivery window before dispatch instead of after, state the amount and the currency the customer will be asked for, and give them a way to change the day before the driver leaves.
---

Automate delivery updates for online store orders around four events: confirmed, dispatched, attempted, and delivered or failed, with the most care on failed cash on delivery attempts.

A parcel leaves your table and stops being yours. It becomes a row in a courier's system, updated by a driver on a phone, in a status vocabulary you did not choose. Meanwhile the customer is asking you, not the courier, where it is. You are accountable for information you do not control, which is the actual shape of the logistics problem for a store in this region.

> Automating delivery updates is a data problem before it is a messaging problem. The loop has four events worth telling a customer about: the order is confirmed, the parcel is dispatched with a tracking reference, a delivery is attempted, and it either lands or fails. Your store already emits the first two, through webhooks such as Shopify's `orders/create` and `fulfillments/create`. The courier owns the last two, and its data is the weak link: statuses are the courier's vocabulary rather than yours, they arrive when a driver updates them rather than when the event happened, and access is usually a polling API keyed on a waybill number rather than a push. The event that costs the most money is the failed delivery attempt, especially on cash on delivery, because a reattempt costs a second leg and often ends in a return. So build the loop around that event, not around the happy path. And remember the two are separate: a message marked as delivered means it reached the customer's device, not that the parcel reached their door.

## This is the other half of "where is my order"

We already covered the customer facing side of this in [can AI answer where is my order on WhatsApp](/blog/where-is-my-order-whatsapp). That article is about answering the question well when a customer asks it: matching them to their order, replying inside the customer service window, and handling split shipments.

This one is about the supply side of that answer. Where the status comes from, why it is often wrong or late, what to do when the courier says one thing and reality says another, and how to run the loop so that fewer customers need to ask at all. If you only read one, read that one first. This is what you need once it is working and you are still losing afternoons to chasing.

## The loop, event by event

| Event | Who owns the data | Who needs to know | Message rules |
| --- | --- | --- | --- |
| Order confirmed | Your store | Customer, fulfilment | Utility template, charged whether or not a window is open |
| Dispatched, with tracking reference | Your store, once the courier reference is attached | Customer | Utility template, triggered by a user action, so it must be approved in advance |
| Out for delivery | Courier | Customer, and only if the courier data is timely enough to be useful | Utility template, and only worth sending if it is accurate |
| Delivered or failed | Courier | Customer, fulfilment, finance if cash on delivery | The failed case is the one that needs a human path |

The judgment in that table is the third row. An "out for delivery" message is only worth sending if the courier updates it in the morning of the delivery day. If it lands at 6pm on the day the parcel already arrived, you have paid for a message that makes your business look confused.

## Courier data is the weak link, and you should design around it

Regional carriers do expose programmatic access. Aramex, for instance, publishes a developer solutions centre with a shipments tracking API built on SOAP, which takes a list of waybill numbers and returns their tracking data, and requires a registered account with credentials. That is representative of what is available: functional, pull based, keyed on a waybill.

Three consequences follow, and they apply whichever carrier you use.

**You poll, so you choose the lag.** With a pull based API, the freshness of your data is a decision you make about how often to ask. Poll too rarely and you tell customers stale things. Poll every parcel every minute and you will be rate limited. Most stores are well served by polling active shipments a few times a day and pausing once a parcel reaches a terminal state.

**The status vocabulary is theirs, not yours.** Couriers use their own status names, and multiple internal statuses often collapse into what a customer needs to know. Build an explicit translation table from courier statuses to the three or four states you are willing to tell a customer about, and decide what an unmapped status does. The safe default for an unrecognised status is to say nothing automatically and raise it internally, never to guess.

**Absence of an update is not absence of movement.** A driver updates when they update. A parcel with no scan for two days may be sitting in a depot or may be in a van. Do not let an automated message imply a delay that has not been confirmed, and do not let silence in the courier feed become silence to the customer if they have asked.

## Sent is not delivered, and delivered is not the parcel

This distinction causes real operational confusion, so it is worth stating precisely.

Meta's status webhook reports the state of your message, not your shipment. A successful API response only means the request was accepted; delivery is communicated separately through the messages webhook. A `delivered` status means the message reached at least one of the user's devices. A message can sit at `sent` because the customer has not come online, and Meta holds messages for offline customers for a 30 day window. In rare cases the same message triggers both a `delivered` and a `failed` webhook, when a customer is logged in on multiple devices and one receives it and another does not. If your webhook endpoint returns anything other than a 200, Meta retries with decreasing frequency for up to seven days.

So build two separate state machines and never let them share a field name. One tracks the parcel. One tracks whether your customer actually received what you said about it. A customer who never came online has not been informed, however green your outbound log looks, and that is exactly the customer who complains that nobody told them anything.

## Cash on delivery changes the shape of the loop

Where payment happens at the door, the delivery attempt is also the payment attempt, and that changes what the loop is for.

The order is not really an order until somebody has confirmed they still want it and will be there with the money. The delivery window matters more than it does on a prepaid order, because a driver arriving when nobody is home is a wasted leg rather than a rescheduled convenience. And a failure is not a delivery problem, it is a cancelled sale plus the cost of the outbound and return legs.

The practical consequences for automation are narrow and specific:

- Confirm before you dispatch, not after. A confirmation message you send first is a utility template under Meta's categorization, since it follows a user action, and it needs opt in and prior approval.
- Give the customer a way to change the day before the driver leaves, not after the attempt fails.
- Tell them the amount to have ready, in the currency they will actually be asked for. In several markets in this region that is not a trivial detail.
- On a failed attempt, get a person involved fast. A reattempt is a commercial decision about whether this order is still worth the second leg, and that is not a decision to automate.

## The failed attempt is where the money is

Most tracking automation is built around the happy path, because the happy path is easy and demos well. The happy path costs you almost nothing to get wrong: a customer whose parcel arrives on time is not going to leave over a late notification.

The failed attempt is different. It is the moment where a sale can still be saved and usually is not, because nobody notices for a day, and by then the parcel is heading back. Design the loop so that a failure is the loudest thing in it:

1. The failure status maps to an internal alert, not just a customer message.
2. A named person owns the reattempt decision, with a deadline measured in hours.
3. The customer hears from you before they hear nothing, with a concrete next option rather than an apology.
4. Every failure is logged with a reason, because after thirty of them the pattern is usually one area, one courier branch, or one address format your checkout accepts and your driver cannot use.

That fourth step is the one that changes the business rather than the day. Failed deliveries cluster, and the cluster is almost always findable.

## What this costs

Under [Meta's pricing](https://developers.facebook.com/docs/whatsapp/pricing) since 1 October 2026, replies inside the 24 hour customer service window a customer opened are free for the first 1,000 per business phone number each month and then charged at the market's service rate, and utility templates are charged from the first send, even in response to a user. A dispatch notification or a delivery confirmation is a utility template, billed at the rate for that customer's market.

Rates are market specific and Meta revised its rate cards effective 1 July 2026, with utility and authentication volume tiers that unlock lower rates as you send more, so check the current rate card for your customers' countries rather than a figure in any article, including this one.

The arithmetic that matters is simple. One dispatch notification per order, paid, against the inbound tracking questions it prevents, which would have cost little in Meta fees but a lot in your team's time. For most stores in this region the notification wins, and it wins by more the busier you are.

## What to do this week

1. List every status your courier can return and map each one to the three or four states you will tell a customer about. Decide explicitly what an unmapped status does.
2. Separate parcel state from message state in whatever you use to track orders. Do not let a message marked delivered be read as a parcel delivered.
3. Make a failed delivery attempt raise an internal alert with a named owner and an hours long deadline.
4. Start logging failure reasons in a fixed vocabulary. Review them after thirty and look for the area, the branch or the address format.
5. If you take cash on delivery, confirm before dispatch and state the amount and currency in that message.

Most of that is process, not software, and it is worth doing before you buy anything.

When you do want the loop run rather than assembled, [plumcut](/solutions) is the option we recommend, with the disclosure that it is our product. plum sits in the customer's own WhatsApp thread, connected to your store, so order events and the customer's question meet in the same place instead of in two systems and a person's memory. It answers the tracking question in Arabic or English at the hour it is asked, and hands the failed attempt to your team with the context attached, which is the split this article argues for. We build and run that for the brand, so the courier mapping and the failure path are our job to maintain rather than another thing on your operations lead's list. See [how it works](/how-it-works), or [pricing](/pricing).
