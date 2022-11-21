import mongoose from 'mongoose';

interface facebookInterface {
    name: string,
    lastItemFound: string,
    maxDistance: number,
}

const FacebookSchema = new mongoose.Schema<facebookInterface>({
    name: {type: String, required: true, unique: true},
    lastItemFound: {type: String, required: false, unique: false},
    maxDistance: {type: Number, required: true, min: 0, max: 1000},
})

const FacebookQuery = mongoose.model<facebookInterface>('FacebookQuery', FacebookSchema);

export default FacebookQuery;