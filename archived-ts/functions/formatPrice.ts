const formatter = new Intl.NumberFormat("en-AU", {
  style: "currency",
  currency: "AUD",
});
export function formatPrice(price: number) {
  return formatter.format(price);
}
