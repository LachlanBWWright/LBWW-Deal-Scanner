export default function (error: Error, name: string) {
  console.error("An error has occurred in " + name + ":");
  console.error(error);
  //An option of emitting the error to a discord channel may be added in future
}
