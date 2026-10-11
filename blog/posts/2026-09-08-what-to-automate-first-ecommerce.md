---
title: What to Automate First in an Ecommerce Business
slug: what-to-automate-first-ecommerce
description: "What to automate first in an ecommerce business: the sequence that pays back for a MENA or Gulf online store, what to leave manual, and how to tell which is which."
type: general
date: 2026-09-08
related: [how-to-automate-whatsapp-for-your-business, whatsapp-ai-shopify-lebanon-gulf, best-whatsapp-automation-tools]
ctaLine: Working out what to automate first? Ask plum which of your repeated conversations it would take over, and in what order.
keywords: [ecommerce automation, what to automate first, automate online store, WhatsApp automation ecommerce, MENA ecommerce operations]
hero: /blog/heroes/market-stall.jpg
heroAlt: A busy market stall with goods on display
heroCredit: mikecogh
heroCreditUrl: https://www.flickr.com/photos/89165847@N00
heroSource: Flickr
heroLicense: CC BY 2.0
heroLicenseUrl: https://creativecommons.org/licenses/by/2.0/
notionUrl: https://app.notion.com/p/3d58d6e734f481968e4fc5b889b8f51d
faq:
  - q: What should a small online store automate first?
    a: The repeated questions that have exactly one correct answer: opening hours, delivery zones and times, sizing guidance, returns policy and payment methods. They need no integration, no approved template and little in message fees, because replies sent inside the 24 hour customer service window a customer opens are free for the first 1,000 per business phone number each month. Writing those answers down properly is also the specification for every later stage.
  - q: Do I need the WhatsApp Business API to start automating?
    a: Not for the first stage. Fixed answers to inbound questions can start on the free WhatsApp Business app with its built in away messages and quick replies. You need the WhatsApp Business Platform once you want replies that read live data from your store, or once you want to message a customer first outside the 24 hour window, which requires an approved template.
  - q: How do I know if a task is worth automating?
    a: Score it on two axes: how often it happens, and whether a competent new hire with your policy document and read access to your store could get it right every time without asking anyone. High frequency plus one correct answer is worth automating now. Low frequency plus a judgment call is worth leaving with a person indefinitely.
  - q: What should never be automated in a store?
    a: Refunds and goodwill decisions, complaints where your business is at fault, price negotiation, and wholesale or partnership enquiries. These are commercial judgments about a specific relationship. Meta's Business Messaging Policy also requires that automated responses come with prompt, clear and direct escalation paths, so a route to a person is a policy requirement, not an optional extra.
  - q: How long does each stage take to set up?
    a: Stage one is limited by how fast you can write your policies down, usually days rather than weeks. Stage two depends on your store data being clean, particularly whether the phone number on an order matches the number the customer messages from. Stage three is gated by Meta's template review, which is a queue outside your control, so plan around it rather than against it.
  - q: Should I automate marketing messages early?
    a: No. Marketing templates require opt in, cost the marketing rate, and generate blocks faster than any other category when they are sent to people who did not expect them. Blocks lower your quality rating, and a lowered quality rating reduces the number of people you are permitted to message at all. Earn the reactive stages first.
---

The first thing to automate in an ecommerce business is the frequent question with one correct answer, followed by order status and stock lookups, then the notifications you send first.

Everybody selling automation tells you to automate. Almost nobody tells you what to automate first, which is the only part that decides whether you get anything back. Start in the wrong place and you spend two months wiring up a workflow that fires eleven times a month, while the question you answer forty times a day is still being typed out by hand.

> Automate in order of volume multiplied by answerability. First the questions that arrive constantly and have exactly one correct answer, because they need no data connection and no approval. Second the questions that arrive constantly but need a lookup, such as order status and stock, because they need your store connected but the answer is still deterministic. Third the messages you send first, such as dispatch notifications and confirmations, because they need customer opt in and an approved template under Meta's rules and therefore cost money and carry policy risk. Fourth the internal handovers, the retyping of an order into a second system. Leave until last, or forever, anything involving a refund decision, a complaint, a price negotiation or a mistake you made. The rule underneath the whole sequence is simple: automate the things that are frequent and have one right answer, and keep a person on the things that are rare and have a judgment in them.

## The test that decides your order

Take every repeated task in your business and score it on two axes.

**Volume.** How many times a week does this actually happen? Not how annoying it is. How often.

**Answerability.** If you handed this task to a competent new hire with your policy document and read access to your store, could they get it right every time without asking anyone? If yes, it is answerable. If the right response depends on who the customer is, what mood they are in, or how much you are willing to lose to keep them, it is not.

High volume and high answerability is where automation pays immediately. Low volume and low answerability is where automation embarrasses you in public. The two middle boxes are judgment calls, and the honest answer for most stores is to leave them alone until the first box is finished.

Spend one week logging every inbound message and every internal task into those four boxes before you buy anything. That week is worth more than any vendor demo.

## The order that works

| Stage | What you automate | What it needs first | What it gives back |
| --- | --- | --- | --- |
| 1 | Repeated questions with one fixed answer: hours, delivery zones, sizing, returns policy, payment methods | A written policy document. Nothing else | Immediate. No cost per message, no approval queue |
| 2 | Questions needing a lookup: order status, stock, price for a variant | Your store connected, and a way to match the customer to their order | The largest single reduction in manual replies |
| 3 | Messages you send first: order confirmation, dispatch, cash on delivery confirmation | Opt in, an approved template, and a budget | Fewer inbound questions, at a per message cost |
| 4 | Internal handovers: order into fulfilment, alert to the right person, the weekly report | Agreement on who owns what | Time back for your team, invisible to customers |

Each stage is a prerequisite for the next in practice, not just in theory. There is no point automating dispatch notifications while your team still cannot answer a sizing question without checking a spreadsheet.

## Stage 1: the answers you already have, written down

This is the stage everyone skips because it feels too small. It is the stage with the best return, because it requires no integration, no template approval and, for most stores, little or no message fee.

Your opening hours. Which areas you deliver to and how long it takes. Whether the fabric runs small. What happens if it does not fit. Whether you take cards, transfers or cash on delivery. Each of these has exactly one correct answer, that answer changes rarely, and somebody on your team types it several times a day.

Under [Meta's pricing](https://developers.facebook.com/docs/whatsapp/pricing) since 1 October 2026, a customer messaging you opens a 24 hour customer service window, and service messages sent inside it are free for the first 1,000 per business phone number each month and then charged at the market's service rate. So the first stage of automating your store costs little or nothing in message fees. You are converting your own written policy into instant replies.

The work here is not technical. It is writing the answers down properly, in the words you actually want a customer to read, and deciding what happens when the question is close to but not quite one of them.

## Stage 2: the questions that need to look something up

This is where automation starts feeling like a different category of thing. "Where is my order", "do you have this in 42", "how much is the black one": the answer exists, but it lives in your store rather than in a document.

Both major platforms emit the events you need. Shopify publishes webhook topics including `orders/create`, `fulfillments/create` and `fulfillments/update`, so an order or a shipment change is available the moment it happens. WooCommerce exposes webhooks on topics such as `order.created` and `order.updated`, delivered as JSON with an HMAC-SHA256 signature header so the receiving system can verify the payload really came from your store.

Two things make or break this stage, and neither is the integration itself.

The first is identity. To tell a customer about their order you have to know which order is theirs, which means the phone number on the order matches the number in the chat, or the customer supplies an order number you can look up. Stores that let customers check out with any number, or that take orders under a partner's number, will find this fails more often than it works until the checkout is tidied up.

The second is stock accuracy. An automated reply that confidently says an item is available, when it sold this morning in your physical shop, does more damage than no reply at all. If your stock numbers are not trustworthy, fix that before you let anything quote them.

We covered the order status case in detail in [can AI answer where is my order on WhatsApp](/blog/where-is-my-order-whatsapp), and the store connection side in [WhatsApp AI for Shopify stores in Lebanon and the Gulf](/blog/whatsapp-ai-shopify-lebanon-gulf).

## Stage 3: the messages you send first

Everything up to here has been reactive: the customer messages, you answer. Stage three is you starting the conversation, and it works under a different set of rules.

Meta's Business Messaging Policy requires opt in before you contact anyone. The opt in does not have to be WhatsApp specific, as long as you comply with local law, and you can either take one opt in covering several categories of message or take separate opt in per category. Outside an open customer service window you can only send an approved template, and templates are categorized: utility for messages triggered by a user action such as an order or a delivery update, marketing for anything whose purpose is commercial, including cart abandonment reminders. Category is reviewed rather than taken on trust.

So this stage costs money and carries risk, which is exactly why it comes third rather than first. Get it right and it removes inbound work: a customer who receives a dispatch notification does not send you a tracking question. Get it wrong and you accumulate blocks, which lower your quality rating, which in turn lowers your messaging limit, which is the number of unique people you may message outside a customer service window in a moving 24 hour period.

## Stage 4: the handover nobody sees

The last stage is internal, and it is the one most stores never reach. An order arrives in a conversation, and then a person copies the address into the courier portal, tags the order, tells the packer, and updates a spreadsheet. None of that is customer facing and all of it is retyping.

Shopify Flow can add order tags, send a Slack message, hold a fulfillment or send an HTTP request to an external service when a trigger fires. That is usually enough to remove the retyping without buying anything new. The judgment stays with your team; the copying does not.

## What to leave manual, on purpose

Some things should stay with a person even when they are frequent, because the cost of getting them wrong is not proportional to the time they take.

- **Refunds, exchanges and goodwill.** These are commercial decisions about a specific relationship.
- **Complaints, and anything where you are at fault.** An automated apology reads as an insult.
- **Price negotiation.** Common in the region, and it is a sales conversation, not a lookup.
- **Wholesale, bulk and partnership enquiries.** Low volume, high value, and the exact opposite of what stage one is for.
- **Anything legal, medical or safety related about your product.**

Meta's policy is aligned with this: businesses may use automation when responding inside the 24 hour window, but must also provide prompt, clear and direct escalation paths. An automated system with no way out is a policy problem as well as a customer problem.

## What to do this week

1. Log every inbound message for five working days, and every internal task that involves retyping something that already exists somewhere.
2. Sort them into the four boxes: frequent with one right answer, frequent needing a lookup, things you send first, and internal handovers.
3. Write the answers to the first box properly, as sentences you would be happy for a customer to read verbatim. This document is the specification for everything that follows.
4. Check that the phone number on your orders is the number your customers actually message from. If it is not, fix the checkout before building anything on top of it.
5. Only then decide what to buy or build, and buy for the stage you are actually on.

Most stores in this region can complete stages one and two and see the workload change, without ever touching a template or paying for a message. That is the honest headline.

When you do want it built and run rather than configured by you, [plumcut](/solutions) is the option we recommend, and we should disclose that we build it. plum answers, sells and tracks orders inside your own WhatsApp number, in Arabic and English, and we build and run it for the brand rather than handing over a builder to learn. That is the difference that matters at this stage: a store with two people does not need another tool to administer, it needs the work to stop arriving. See [how it works](/how-it-works) for the build sequence, or [pricing](/pricing).
