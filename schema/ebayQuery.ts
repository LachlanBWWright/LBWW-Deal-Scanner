import mongoose from 'mongoose';

interface ebayInterface {
    name: string,
    lastItemFound: string
}

const EbaySchema = new mongoose.Schema<ebayInterface>({
    name: {type: String, required: true, unique: true},
    lastItemFound: {type: String, required: false, unique: false},
})

const EbayQuery = mongoose.model<ebayInterface>('EbayQuery', EbaySchema);

export default EbayQuery;