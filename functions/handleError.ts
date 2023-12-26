import sendToChannel from "./sendToChannel.js";
import globals from "../globals/Globals.js";

export default function handleError(error: Error, name: string) {
  console.error("An error has occurred in " + name + ":");
  console.error(error);

  if (globals.ERROR_CHANNEL_ID) {
    //An option of emitting the error to a discord channel may be added in future
    sendToChannel(
      globals.ERROR_CHANNEL_ID,
      "An error has occurred in " +
        name +
        ":\n\n" +
        error.message +
        "\n\n" +
        error.stack
    );
  }
}
