import mongoose from 'mongoose';

interface tradeItInterface {
    name: String,
    maxPrice: Number,
    minFloat: Number,
    maxFloat: Number,
    found: Boolean
}

const TradeItItemSchema = new mongoose.Schema<tradeItInterface>({
    name: {type: String, required: true, unique: true},
    maxPrice: {type: Number, required: true, min: 0, max: 100},
    minFloat: {type: Number, required: true, min: 0, max: 1, default: 0},
    maxFloat: {type: Number, required: true, min:0, max: 1},
    found: {type: Boolean, required: true, default: false}
})

const TradeItItem = mongoose.model<tradeItInterface>('TradeItItem', TradeItItemSchema);

export default TradeItItem;