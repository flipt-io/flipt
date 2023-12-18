import { useState } from 'react';
import Modal from '~/components/Modal';

function LoginForm() {
  const [formData, setFormData] = useState({
    username: '',
    password: ''
  });
  const [modalOpen, setModalOpen] = useState(false);
  const [loginError, setLoginError] = useState('');
  const [loginSuccess, setLoginSuccess] = useState('');

  const handleInputChange = (e: { target: { name: any; value: any } }) => {
    const { name, value } = e.target;
    setFormData({ ...formData, [name]: value });
  };

  const handleSubmit = async (e: { preventDefault: () => void }) => {
    e.preventDefault();
    setLoginError('');
    setLoginSuccess('');

    try {
      const response = await fetch(
        'http://localhost:8080/auth/v1/method/basic/login',
        {
          method: 'POST',
          headers: {
            'Content-Type': 'application/json'
          },
          body: JSON.stringify(formData)
        }
      );

      if (response.ok) {
        setLoginSuccess('Success!🎉');
      } else {
        const errorData = await response.json();
        setLoginError(errorData.message || 'Failed to log in.');
      }
    } catch (error) {
      console.error('Network or server error:', error);
      setLoginError('Network error or server is unreachable.');
    }
  };

  const closeModal = () => {
    setLoginError('');
    setLoginSuccess('');
    setFormData({ username: '', password: '' });
    setModalOpen(false);
  };

  const isFormIncomplete = !formData.username || !formData.password;

  return (
    <>
      <button
        onClick={() => setModalOpen(true)}
        className="text-white bg-purple-500 rounded px-4 py-2 font-bold hover:bg-purple-700"
      >
        Login
      </button>

      <Modal open={modalOpen} setOpen={setModalOpen}>
        <form onSubmit={handleSubmit} className="p-4">
          <div className="flex items-center justify-between">
            <h2 className="mb-4 text-2xl font-semibold">Login</h2>
          </div>
          <div className="mb-4">
            <label htmlFor="username" className="text-gray-700 block font-bold">
              Username
            </label>
            <input
              type="text"
              id="username"
              name="username"
              value={formData.username}
              onChange={handleInputChange}
              required
              className="border-purple-400 w-full rounded border p-2 focus:border-purple-500 focus:outline-none"
            />
          </div>
          <div className="mb-4">
            <label htmlFor="password" className="text-gray-700 block font-bold">
              Password
            </label>
            <input
              type="password"
              id="password"
              name="password"
              value={formData.password}
              onChange={handleInputChange}
              required
              className="border-purple-400 w-full rounded border p-2 focus:border-purple-500 focus:outline-none"
            />
          </div>
          {loginError && <div className="text-red-500 mb-3">{loginError}</div>}
          {loginSuccess && (
            <div className="text-green-500 mb-3">{loginSuccess}</div>
          )}
          <div className="flex justify-center gap-4 text-center">
            <button
              type="submit"
              className={`text-white rounded px-4 py-2 font-bold ${
                isFormIncomplete
                  ? 'bg-purple-300 cursor-not-allowed hover:bg-purple-400'
                  : 'bg-purple-500 hover:bg-purple-700'
              }`}
              disabled={isFormIncomplete}
              title={
                isFormIncomplete
                  ? 'Fill in username and password to submit'
                  : ''
              }
            >
              Login
            </button>

            <button
              type="button"
              className="hover:bg-white-700 text-purple-500 border-purple-500 rounded border px-4 py-2 font-bold hover:border-purple-700"
              onClick={closeModal}
            >
              Cancel
            </button>
          </div>
        </form>
      </Modal>
    </>
  );
}

export default LoginForm;
