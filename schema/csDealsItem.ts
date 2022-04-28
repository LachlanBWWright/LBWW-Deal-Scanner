import mongoose from 'mongoose';

interface csDealsInterface {
    name: String,
    maxPrice: Number,
    minFloat: Number,
    maxFloat: Number
}

const CsDealsItemSchema = new mongoose.Schema<csDealsInterface>({
    name: {type: String, required: true, unique: true},
    maxPrice: {type: Number, required: true, min: 0, max: 100},
    minFloat: {type: Number, required: true, min: 0, max: 1, default: 0},
    maxFloat: {type: Number, required: true, min:0, max: 1}
})

const CsDealsItem = mongoose.model<csDealsInterface>('CsDealsItem', CsDealsItemSchema);

export default CsDealsItem;