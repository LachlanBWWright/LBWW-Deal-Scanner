import mongoose from 'mongoose';

interface csMarketInterface {
    name: String,
    displayUrl: String,
    maxPrice: Number,
    maxFloat: Number,
    lastFloat: Number,
    found: Boolean
}

const CsMarketItemSchema = new mongoose.Schema<csMarketInterface>({
    name: {type: String, required: true, unique: true},
    displayUrl: {type: String, required: false},
    maxPrice: {type: Number, required: true, min: 0, max: 100},
    maxFloat: {type: Number, required: true, min:0, max: 1}
})

const CsMarketItem = mongoose.model<csMarketInterface>('CsMarketItem', CsMarketItemSchema);

export default CsMarketItem;