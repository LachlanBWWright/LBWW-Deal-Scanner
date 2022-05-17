import mongoose from 'mongoose';

interface steamQueryInterface {
    name: String,
    maxPrice: Number,
    lastPrice: Number
}

const SteamQuerySchema = new mongoose.Schema<steamQueryInterface>({
    name: {type: String, required: true, unique: true},
    maxPrice: {type: Number, required: true, min: 0, max: 1000},
    lastPrice: {type: Number, required: false, default: 0}
})

const SteamQuery = mongoose.model<steamQueryInterface>('SteamQuery', SteamQuerySchema);

export default SteamQuery;