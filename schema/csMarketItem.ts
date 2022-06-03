import mongoose from 'mongoose';

interface csMarketInterface {
    name: string,
    displayUrl: string,
    maxPrice: number,
    maxFloat: number,
}

const CsMarketItemSchema = new mongoose.Schema<csMarketInterface>({
    name: {type: String, required: true, unique: true},
    displayUrl: {type: String, required: true, unique: true},
    maxPrice: {type: Number, required: true, min: 0, max: 250},
    maxFloat: {type: Number, required: true, min:0, max: 1}
})

const CsMarketItem = mongoose.model<csMarketInterface>('CsMarketItem', CsMarketItemSchema);

export default CsMarketItem;