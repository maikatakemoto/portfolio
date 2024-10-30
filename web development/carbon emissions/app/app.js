import express from 'express';
const app = express();
const port = 9000;

app.set('views', './app/views');
app.set('view engine', 'ejs');

app.use(express.static('public'));
app.use(express.urlencoded({ extended: true }));

app.get('/', (req, res) => {
    res.render('index', { username: 'Guest' });
});

app.get('/about', (req, res) => {
    const { username } = req.body;
    res.render('about', { username }); 
});

app.get('/calculator', (req, res) => {
    const { username } = req.body;
    res.render('calculator', { username }); 
});

app.listen(port, () => {
    console.log(`Server is running at http://localhost:${port}`);
});
