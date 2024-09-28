import axios from 'axios';

export const seed = (user: string) => axios.get('/seed?user=' + user)
    .then(response => response.data as number);
