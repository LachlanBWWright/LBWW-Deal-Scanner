import mongoose from 'mongoose';

interface csMarketInterface {
    name: String,
    maxPrice: Number,
    maxFloat: Number,
    lastFloat: Number
}

const CsMarketItemSchema = new mongoose.Schema<csMarketInterface>({
    name: {type: String, required: true, unique: true},
    maxPrice: {type: Number, required: true, min: 0, max: 100},
    maxFloat: {type: Number, required: true, min:0, max: 1},
    lastFloat: {type: Number, default: 0}
})

const CsMarketItem = mongoose.model<csMarketInterface>('CsMarketItem', CsMarketItemSchema);

export default CsMarketItem;