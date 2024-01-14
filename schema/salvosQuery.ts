import mongoose from "mongoose";

interface salvosInterface {
  name: string;
  minPrice: number;
  maxPrice: number;
  lastItemFound: string;
}

const SalvosSchema = new mongoose.Schema<salvosInterface>({
  name: { type: String, required: true, unique: true },
  minPrice: { type: Number, required: false, unique: false },
  maxPrice: { type: Number, required: false, unique: false },
  lastItemFound: { type: String, required: false, unique: false },
});

const SalvosQuery = mongoose.model<salvosInterface>(
  "SalvosQuery",
  SalvosSchema
);

export default SalvosQuery;
