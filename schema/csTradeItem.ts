import mongoose from "mongoose";

interface csTradeInterface {
  name: string;
  maxPrice: number;
  minFloat: number;
  maxFloat: number;
  found: boolean;
}

const CsTradeItemSchema = new mongoose.Schema<csTradeInterface>({
  name: { type: String, required: true, unique: true },
  maxPrice: { type: Number, required: true, min: 0, max: 100 },
  minFloat: { type: Number, required: true, min: 0, max: 1, default: 0 },
  maxFloat: { type: Number, required: true, min: 0, max: 1 },
  found: { type: Boolean, required: true, default: false },
});

const CsTradeItem = mongoose.model<csTradeInterface>(
  "CsTradeItem",
  CsTradeItemSchema,
);

export default CsTradeItem;
