const standardPreludes = [
  "Listen up, Jack,",
  "My fellow Americans,",
  "Folks,",
  "Here's the deal,",
];

const rarePreludes = ["Look, fat,"];

const rareNotificationPreludes = ["This is a big fu- ...uh... flippable deal,"];

const negativePreludes = [
  "Malarkey,",
  "This is a bunch of malarkey,",
  "Uh-oh, Jack,",
];

const rareNegativePreludes = ["Stupid son of a bitch,"];

export function getNotificationPrelude() {
  return getResponseMessage(standardPreludes, rareNotificationPreludes);
}

export function getResponsePrelude() {
  return getResponseMessage(standardPreludes, rarePreludes);
}

export function getFailurePrelude() {
  return getResponseMessage(negativePreludes, rareNegativePreludes);
}

function getResponseMessage(
  standardMessages: string[],
  rareMessages: string[],
) {
  if (Math.random() < 0.01) {
    return rareMessages[Math.floor(Math.random() * rareMessages.length)];
  }
  return standardMessages[Math.floor(Math.random() * standardMessages.length)];
}
