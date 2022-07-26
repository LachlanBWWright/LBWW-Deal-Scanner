import mongoose from 'mongoose';

interface gumtreeInterface {
    name: string,
    lastItemFound: string
    maxPrice: number
}

const GumtreeSchema = new mongoose.Schema<gumtreeInterface>({
    name: {type: String, required: true, unique: true},
    lastItemFound: {type: String, required: false, unique: false},
    maxPrice: {type: Number, required: true, min: 0, max: 1000}
})

const GumtreeQuery = mongoose.model<gumtreeInterface>('GumtreeQuery', GumtreeSchema);

export default GumtreeQuery;