---
title: Can AI Answer Where Is My Order on WhatsApp in 2026?
slug: where-is-my-order-whatsapp
description: "Where is my order on WhatsApp: what it takes to answer the most common question in commerce automatically, what it costs, and where it breaks."
type: general
date: 2026-09-06
related: [whatsapp-ai-shopify-lebanon-gulf, whatsapp-business-api-cost, how-to-automate-whatsapp-for-your-business]
ctaLine: Spending too much time answering delivery questions? Ask plum how a managed WhatsApp workflow can connect customers to the right order information.
keywords: [WhatsApp order status updates, where is my order automation, WhatsApp shipping notification, WhatsApp utility template, Shopify WhatsApp order tracking]
hero: /blog/heroes/sealed-box.jpg
heroAlt: A sealed cardboard box ready to ship
heroCredit: public.resource.org
heroCreditUrl: https://www.flickr.com/photos/8212496@N06
heroSource: Flickr
heroLicense: CC0 1.0
heroLicenseUrl: https://creativecommons.org/publicdomain/zero/1.0/
notionUrl: https://app.notion.com/p/3d38d6e734f481b5bbbac92c3fe28f9d
faq:
  - q: Does answering "where is my order" on WhatsApp cost anything?
    a: Not much when you are replying. A customer message opens a 24 hour customer service window, and since 1 October 2026 Meta charges for service messages sent inside it only after the first 1,000 per business phone number each month. What costs money is messaging first, for example a dispatch notification sent when no window is open, which is billed as a utility template at the rate for that customer's country.
  - q: Can I send tracking updates before the customer asks?
    a: Yes, with an approved utility template and the customer's opt in. Sent outside an open customer service window it is charged at the utility rate. Sent inside one it is also charged at the utility rate, since 1 October 2026. Most brands find that one dispatch notification per order pays for itself by removing the question entirely.
  - q: What happens when one order ships in two boxes?
    a: A Shopify fulfilment carries a single tracking number, so a split shipment is two fulfilments and two events, while your customer is still thinking about one order. Decide the wording in advance. Say which items have shipped and which are still coming, because a plain "your order is on its way" gets discovered at the door.
  - q: Do I need the WhatsApp Business API to automate order status?
    a: For reactive answers to customers who message you, not necessarily. For proactive updates you send first, yes. The free WhatsApp Business app cannot send a template to a customer outside the 24 hour window and cannot read your store data, so anything that fires automatically on a fulfilment event needs the WhatsApp Business Platform.
  - q: How quickly does an order status reply need to arrive?
    a: Fast enough to land while the customer is still in the app, which in practice means seconds rather than minutes. Speed also keeps the reply cheap, since it happens inside the customer service window the customer just opened, where the first 1,000 service replies per number each month are free. A reply that arrives the next morning costs you the same as no reply.
---

AI can answer where is my order on WhatsApp in 2026 when your store sends tracking events and the chat can be matched to the order by phone number or order number.

Roughly the same message arrives every day, in a dozen spellings. "Where is my order." "Any update on my parcel." "Wein el order?" It is the highest volume question in commerce, it carries no new information, and somebody on your team is typing the answer by hand while a customer with money in their hand waits behind it in the queue.

> Order status is the one customer question that automates cleanly, because the answer already exists inside your store and only needs fetching. Three things have to be true. Your store has to emit the event: Shopify fires `fulfillments/create`, `fulfillments/update` and `orders/fulfilled` webhooks carrying the tracking number, so the status is available the moment it changes. The message has to be matched to the right customer, which means the phone number on the order and the number in the chat agree, or the customer gives you an order number to look up. And the reply has to fit Meta's rules. A reply sent inside the 24 hour customer service window needs no template, and since 1 October 2026 the first 1,000 such replies per business phone number each month are free. A status update you send first is a utility template that must be approved in advance, and it is charged at the utility rate whether or not a window is open. Answer fast and the answer is usually cheap.

## "Where is my order" is three questions wearing one coat

The same seven words mean different things depending on when they arrive, and an automation that returns one canned reply to all three feels broken even when the data is correct.

| When it arrives | What they actually want | Where the answer lives |
| --- | --- | --- |
| An hour after paying | Proof the order exists and the payment landed | Your store, at order creation |
| Two days later | Has it shipped, and what is the tracking number | The fulfilment record |
| Five days later | When will it arrive, and can I change the address | The courier, and often only by phone |
| After the promised date | Reassurance, or a refund conversation | Nowhere automatic, this one is human |

The first two automate completely. The third automates partially, because most couriers give you a status and a scan location, not a promise. The fourth should never be automated, and the useful design decision is making sure it reaches a person quickly instead of looping the customer through the same status line a third time.

## What your store has to hand over

On Shopify the plumbing is the easy part, and it is worth knowing the shape of it before you buy anything.

Shopify exposes webhook topics including `fulfillments/create`, `fulfillments/update` and `orders/fulfilled`. When a fulfilment is created with a tracking number, Shopify can resolve the tracking company and tracking URL automatically from the number format. That gives you a push event rather than a polling loop, which is the difference between a reply that arrives in two seconds and one that arrives when a cron job next runs.

There is one limit that catches people. A Shopify fulfilment carries a single tracking number. If an order ships in two boxes, that is two fulfilments, two tracking numbers and two events, while the customer is still thinking about one order. Any automation you build has to decide what to say when an order is half shipped, because "your order is on its way" is a lie the customer will discover at the door. Say two of three items shipped, and name the ones still coming.

WooCommerce, Salla and Zid all expose comparable order and shipment events. The pattern is the same everywhere: the store knows, the courier knows, and the customer is the last to find out.

## What it costs to answer, honestly

This is where most order status projects get their budget wrong, in both directions.

Meta bills per message delivered, not per conversation, since 1 July 2025. Two rules matter more than the rate card:

- **Service messages are free up to a limit.** When a customer messages you, a 24 hour customer service window opens. Replies inside that window are not templates, and since 1 October 2026 [Meta](https://developers.facebook.com/docs/whatsapp/pricing) charges for them only after the first 1,000 per business phone number each month. A customer asking "where is my order" has just opened that window themselves.
- **Utility templates are charged from the first send**, inside or outside an open customer service window, at the utility rate for the recipient's country, since 1 October 2026.

So a reactive answer costs nothing until you pass 1,000 service replies in a month. A proactive shipping notification, the one you send at three in the afternoon when the courier scans the parcel, costs the utility rate, because no window is open. That is still usually the better trade, since the notification you send prevents the question you would otherwise answer, but price it deliberately rather than by accident.

Rates move, and they move regionally. Meta's 1 August 2026 pricing update raised utility and authentication rates in several markets including Kuwait, Morocco and Oman, and a further update to service and utility pricing is scheduled for 1 October 2026. Check the [current rate card](https://developers.facebook.com/documentation/business-messaging/whatsapp/pricing) for your markets before you model anything, and read our breakdown of [what the WhatsApp Business API really costs](/blog/whatsapp-business-api-cost) for the four line items nobody warns you about.

### A rough shape of the trade

For a store shipping 600 orders a month, the arithmetic below is an illustration of the structure of the cost, not a benchmark. Put your own volumes and your own market rate into it.

| Approach | Messages Meta bills | What the customer experiences |
| --- | --- | --- |
| Answer only when asked | Zero, all replies fall inside the window | Fast answers, but every customer still has to ask |
| Notify on dispatch, then answer | One utility template per order | Most people never ask, the rest get a reply inside the service window |
| Notify on dispatch, out for delivery, and delivered | Up to three utility templates per order | Fewer questions again, and a real risk of feeling like spam |

The middle row is where most commerce brands land. The third row earns its keep for high value or fragile goods and irritates everybody else.

## Where this breaks in the Gulf and the Levant

Three regional realities decide whether the automation feels premium or feels cheap.

**Courier data quality.** A tracking number from a regional carrier is not always machine readable, and last mile subcontracting means the scan you receive can be hours behind the driver. If you cannot get a reliable status, say what you do know and the time you knew it, rather than inventing precision you do not have.

**Cash on delivery changes the question.** For a customer paying cash, "where is my order" is often really "when do I need to have the money ready and be at home". A status line that ignores the amount due answers the wrong question. Include the amount and the expected day.

**The message will not arrive in clean English.** Real inbound is dialect, Arabizi and code switching, sometimes in one sentence. A system that only parses English and Modern Standard Arabic will treat "wein sar el order tabaee" as noise and hand it to a human, which puts you back where you started. We go through this properly in [does AI customer service really work in Arabic](/blog/ai-customer-service-arabic).

## What to do this week

1. Pull last month's WhatsApp messages and count how many were tracking questions. That number is the size of the prize, and it is usually larger than anyone guesses.
2. Check whether your store actually emits a fulfilment event with a usable tracking number today. If fulfilments are marked by hand at the end of the day, fix that before buying any automation, because no tool can send an event that never fires.
3. Write the three replies by hand: order received, order shipped, order half shipped. If you cannot write the half shipped one clearly, your automation cannot either.
4. Decide your escalation rule now. Ours is simple: after the promised delivery date, a tracking question goes to a person, not to a status lookup.

Answering "where is my order" is not a chatbot problem, it is a data plumbing problem with a conversation on the end of it. [plum](/solutions) is the AI that sells on WhatsApp, and it reads your store so tracking questions get answered in the same thread where the next sale happens. See [how it works](/how-it-works), or start with [how to automate WhatsApp for your business](/blog/how-to-automate-whatsapp-for-your-business).
