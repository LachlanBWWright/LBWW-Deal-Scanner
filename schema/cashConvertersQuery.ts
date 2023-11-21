import mongoose from "mongoose";

interface cashConvertersInterface {
  name: string;
  lastItemFound: string;
}

const CashConvertersSchema = new mongoose.Schema<cashConvertersInterface>({
  name: { type: String, required: true, unique: true },
  lastItemFound: { type: String, required: false, unique: false },
});

const CashConvertersQuery = mongoose.model<cashConvertersInterface>(
  "CashConvertersQuery",
  CashConvertersSchema,
);

export default CashConvertersQuery;
