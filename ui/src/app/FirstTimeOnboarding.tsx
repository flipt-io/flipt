import { useSelector } from 'react-redux';
import { useNavigate } from 'react-router-dom';
import Onboarding from './Onboarding';
import { selectCompletedOnboarding } from './user/userSlice';
import { useEffect } from 'react';

export default function FirstTimeOnboarding() {
  const completedOnboarding = useSelector(selectCompletedOnboarding);
  const navigate = useNavigate();

  useEffect(() => {
    if (completedOnboarding) {
      navigate('/flags');
      return;
    }
  }, [completedOnboarding, navigate]);

  return <Onboarding firstTime={true} />;
}
