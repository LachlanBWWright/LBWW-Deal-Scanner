import mongoose from 'mongoose';

interface ebayInterface {
    name: string,
    lastItemFound: string
    maxPrice: number
}

const EbaySchema = new mongoose.Schema<ebayInterface>({
    name: {type: String, required: true, unique: true},
    lastItemFound: {type: String, required: false, unique: false},
    maxPrice: {type: Number, required: true, min: 0, max: 1000}
})

const EbayQuery = mongoose.model<ebayInterface>('EbayQuery', EbaySchema);

export default EbayQuery;