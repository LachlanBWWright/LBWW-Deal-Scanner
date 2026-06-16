import client from "../globals/DiscordJSClient.js";

export default function setStatus(statusText: string) {
  client.user?.setPresence({
    status: "online",
    activities: [
      {
        name: statusText,
        type: 4,
      },
    ],
  });
}
