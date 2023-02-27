import mongoose from 'mongoose';

interface steamQueryInterface {
    name: string,
    displayUrl: string,
    maxPrice: number,
    lastPrice: number
}

const SteamQuerySchema = new mongoose.Schema<steamQueryInterface>({
    name: {type: String, required: true, unique: true},
    displayUrl: {type: String, required: false},
    maxPrice: {type: Number, required: true, min: 0, max: 1000},
    lastPrice: {type: Number, required: false, default: 0}
})

const SteamQuery = mongoose.model<steamQueryInterface>('SteamQueries', SteamQuerySchema);

export default SteamQuery;