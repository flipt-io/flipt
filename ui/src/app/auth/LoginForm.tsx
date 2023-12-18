import { useState } from 'react';
import Modal from '~/components/Modal';

function LoginForm() {
  const [formData, setFormData] = useState({
    username: '',
    password: ''
  });

  const [modalOpen, setModalOpen] = useState(false);

  const handleInputChange = (e: { target: { name: any; value: any } }) => {
    const { name, value } = e.target;
    setFormData({ ...formData, [name]: value });
  };

  const handleSubmit = async (e: { preventDefault: () => void }) => {
    e.preventDefault();

    const response = await fetch('/auth/v1/method/basic/login', {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json'
      },
      body: JSON.stringify(formData)
    });

    if (response.ok) {
      // Authentication successful
    } else {
      // Authentication failed
    }
  };

  const openModal = () => {
    setModalOpen(true);
  };

  const closeModal = () => {
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
            <button
              onClick={() => setModalOpen(false)}
              className="text-gray-800 bg-gray-200 rounded-full p-2 hover:bg-gray-300"
            >
              &times;
            </button>
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
              className="border-gray-400 w-full rounded border p-2 focus:border-purple-500 focus:outline-none"
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
              className="border-gray-400 w-full rounded border p-2 focus:border-purple-500 focus:outline-none"
            />
          </div>
          <div className="text-center">
            <button
              type="submit"
              className={`text-white rounded px-4 py-2 font-bold ${
                isFormIncomplete
                  ? 'bg-gray-400 cursor-not-allowed hover:bg-gray-400'
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
          </div>
        </form>
      </Modal>
    </>
  );
}

export default LoginForm;
