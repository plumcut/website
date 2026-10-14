---
title: "AI Chatbot Wrong Answers: Who Catches Them and How Fast"
description: "AI chatbot wrong answers will happen. The five ways a WhatsApp assistant fails, a silent failure we fixed, and how human takeover should really work."
date: 2026-10-14
slug: ai-chatbot-wrong-answers
type: general
keywords: [ai chatbot wrong answers, when ai customer service gets it wrong, chatbot human takeover, whatsapp chatbot escalation, chatbot silent failure]
hero: /blog/heroes/hand-keyboard-single.jpg
heroAlt: One hand resting on a keyboard
heroCreditUrl: https://www.rawpixel.com/image/5921683/photo-image-background-phone-public-domain
heroSource: Rawpixel
heroLicense: CC0 1.0
heroLicenseUrl: https://creativecommons.org/publicdomain/zero/1.0/
related: [automate-customer-service-without-losing-customers, ai-chatbot-customer-data-safety, ai-customer-service-arabic]
ctaLine: Ask plum what it says when it does not know, then ask to see how your team takes the chat over.
faq:
  - q: What happens when an AI chatbot gives a wrong answer?
    a: It depends on how the system around the chatbot is built. A well run assistant answers only from the business's own catalogue and policies, says it will check with the team instead of guessing, and lets staff see every conversation live and take it over in the same WhatsApp thread. The damage from a wrong answer is mostly set by how fast someone sees it and corrects it, and whether the customer has to repeat themselves.
  - q: What is a silent failure in an AI chatbot?
    a: A silent failure is when the customer gets no reply at all, even though the system may have generated one. It is often worse than a wrong answer, because nobody sees an error and the customer simply waits. The fix is monitoring that compares replies generated with replies actually delivered, and alerts someone when they do not match.
  - q: Can a person take over a WhatsApp chatbot conversation?
    a: Yes, if the system is built for it. A good takeover sets a flag on that one conversation so the assistant stops replying, lets a staff member answer in the same WhatsApp thread on the same business number, and hands the conversation back to the assistant when the person is done. The assistant should re-check that flag right before it sends anything, because a takeover can happen while it is still writing.
  - q: What is the WhatsApp 24 hour customer service window?
    a: When a customer messages a business on the WhatsApp Business Platform, a 24 hour customer service window opens, and each new message from the customer resets it. Inside the window the business can send free-form replies. Outside it, the business can only send template messages, which generally need Meta's approval first.
  - q: When should an AI chatbot hand a conversation to a person?
    a: When the customer is angry, asks for a refund, raises a health or legal question, asks the same thing more than once, or asks something the assistant cannot answer from the business's own catalogue and policies. Meta's WhatsApp Business Messaging Policy also requires businesses that automate replies to offer prompt, clear and direct escalation paths to a person.
---

AI chatbot wrong answers happen with every assistant, so what matters is how fast you see one, who steps in, and whether the customer ever notices.

Every vendor demo shows the assistant getting it right. We run one on WhatsApp for brands in Lebanon and the region, so this is the other half: the ways it fails, one failure of our own, told plainly, and what we built so that a mistake costs a minute instead of a customer.

> AI chatbot wrong answers are inevitable, so judge an assistant by its recovery, not its demo. Ground it in the business's own catalogue and policies, and make it say "let me check with the team" instead of guessing. Let staff see every conversation live and take it over in the same WhatsApp thread, with a handback when they are done. Escalate anger, refunds, health and legal questions, and repeated questions to a person. And monitor for silence, because a reply that was written but never delivered is worse than a wrong one: nobody sees an error, and the customer just waits.

## The five ways an AI chatbot gets it wrong

"Wrong answer" covers more than a wrong fact. These are the five failures we design against.

| Failure | What the customer sees | How you catch it |
| --- | --- | --- |
| Confident wrong answer | A wrong price, a wrong stock status, a policy that does not exist | Answer only from live catalogue and written policy; review flagged answers |
| Silent failure | Nothing. No reply at all | Compare replies generated with replies delivered; alert on any gap |
| Double reply | The assistant and a staff member both answer, sometimes differently | A takeover flag per conversation, re-checked right before sending |
| Wrong or missing escalation | Stuck with the assistant when they needed a person, or handed over for nothing | Clear escalation rules, plus a visible route to a person |
| Misread voice note or mixed message | An answer to a question they did not ask | Confirm before acting on anything unclear |

The last row matters more in this region than in most. Customers here send short voice notes and mix English, Arabic and Arabizi in one line, such as "3andkon hayda bl aswad?" ("do you have this in black?"). An assistant that misreads that should ask, not guess. We looked at how well this works today in [whether AI customer service really works in Arabic](/blog/ai-customer-service-arabic).

## Silence is worse than a wrong answer: our own incident

When we added the check that decides whether a human or the assistant should answer a conversation, one step in the workflow read the wrong field. A database lookup ran in between, and the step assumed the assistant's reply was still where it had been before that lookup. It was not, so the step erased the reply before it reached the customer.

For about ninety minutes, one customer sent six messages and got nothing back, although a reply had been written each time.

The fix was small: the step now names the exact source of its answer instead of assuming the field is still there. The lost replies and records were rebuilt from the workflow's own execution history, not guessed.

The lesson was bigger. A wrong answer is visible: the customer replies "that's not right", and someone sees it. Silence is invisible. The assistant thinks it answered, the dashboard looks calm, and the customer concludes nobody is there. So we now monitor for replies that were generated but never delivered, and treat any gap as an incident.

## How human takeover should work

Almost every business we speak to assumes its team can step into the chat at any moment. One wanted several staff accounts on the business number so anyone could jump in. That expectation is right, and the system should be built for it.

- **A per-conversation switch.** Each conversation has a flag, take over or return to assistant, shown on a dashboard. Flipping it affects that one customer, not the whole inbox.
- **One thread, one number.** The customer stays in the same WhatsApp thread with the same business number. They cannot tell who is replying, and they never have to start again.
- **Re-check before sending.** A takeover can land while the assistant is still writing. The assistant checks the flag again right before it sends, and drops its reply if a person has taken over.
- **Auto-release.** If the person goes quiet for about 15 minutes, the conversation returns to the assistant, so chats do not get stranded.
- **No false typing.** The assistant's typing indicator is suppressed during a takeover, so the customer is not shown activity that will not produce a reply.
- **One owner.** Decide which staff member owns a taken-over conversation, so two people do not answer the same customer at once.

This is also policy, not just good service. Meta's [WhatsApp Business Messaging Policy](https://whatsappbusiness.com/policy/) permits automated replies within the customer service window but requires businesses to offer prompt, clear and direct escalation paths to a person. We wrote up the rest of the handover design in [how to automate customer service without losing the customer](/blog/automate-customer-service-without-losing-customers).

## The 24 hour window decides what your team can send

Takeover has a clock on it. On the WhatsApp Business Platform, a customer's message opens a [24 hour customer service window](https://developers.facebook.com/documentation/business-messaging/whatsapp/messages/send-messages), and each new message from them resets it. Inside the window, the business can send free-form replies. Outside it, only template messages can be sent, and templates generally need Meta's approval first.

So a dashboard should show how much of the window is left on every taken-over conversation, and offer an approved template once it expires. A staff member who picks up a chat 23 hours late needs to know they have one hour, not find out when the message fails.

On cost, checked on 14 October 2026: Meta's [pricing page](https://developers.facebook.com/documentation/business-messaging/whatsapp/pricing) says that, from 1 October 2026, each business phone number gets 1,000 free service messages a month, and service messages are charged from the 1,001st. Unused free messages do not roll over. Confirm the current figures for your number before you budget, because Meta changes them on a schedule.

## Guardrails that stop most wrong answers

- **Answer only from the business's own catalogue and policies.** If it is not in the catalogue or the written policy, the assistant does not know it.
- **Say "let me check with the team" instead of guessing.** A short wait for a correct answer beats a fast wrong one.
- **Escalate by rule.** Anger, refund requests, health and legal topics, and the same question asked twice all go to a person.
- **Log everything.** Every conversation, every handover and every failure should be recorded so it can be replayed and fixed, not just apologised for.

## Questions to ask any AI chatbot vendor

1. What does it say when it does not know the answer?
2. Who gets notified when it hands over, and how?
3. Can we see every conversation live?
4. Can staff take over a single conversation and hand it back?
5. How are failures logged, and can you replay what happened?
6. How fast would you hear about a silent failure, a reply that was never delivered?

If a vendor answers the last one with "that doesn't happen", keep asking. It happened to us. For the data side of the same conversation, see [is customer data safe with an AI chatbot](/blog/ai-chatbot-customer-data-safety).

## Where plumcut fits

plumcut builds and runs plum, which answers, sells and books on a brand's WhatsApp, with the team able to step in on any conversation. The design above is the one we run, and the incident above is why the delivery monitoring exists. You can see [how it works](/how-it-works), and what the assistant learns from all those conversations is covered in [WhatsApp conversation analysis](/blog/whatsapp-conversation-analysis).

## What to do this week

1. Send your current assistant, or the one you are considering, five questions it should not be able to answer. Note which ones it guesses.
2. Ask your team how they would take over a conversation right now, and time it.
3. Check whether anyone would know if replies stopped being delivered for an hour.
4. Write down the four topics that must always go to a person, and make sure the assistant has them.
