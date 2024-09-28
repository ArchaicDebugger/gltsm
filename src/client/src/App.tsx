import React, { useCallback } from 'react';
import logo from './logo.svg';
import './App.css';
import { seed } from './actions/seedActions';

function App() {
  const [user, setUser] = React.useState<string>('');
  const onSubmit = useCallback(() => {
    seed(user).then(data => console.log(data));
  }, [user]);

  return (
    <div className="App">
      <header className="App-header">
        <label htmlFor="user">User:</label>
        <input id="user" type="text" value={user} onChange={e => setUser(e.target.value)} />
        <button type="submit" onClick={onSubmit}>Submit</button>
      </header>
    </div>
  );
}

export default App;
