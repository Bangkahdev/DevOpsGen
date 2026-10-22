import React from 'react';
import { createRoot } from 'react-dom/client';
import './styles/site.css';
import Layout from './components/Layout.jsx';
import { ROUTES } from './routes.jsx';

const route = ROUTES.find((r) => r.path === location.pathname.replace(/\/$/, '').replace(/^.*\//, '')) ??
  null;

const rel = route && route.path === '' ? '' : '../';
const page = route ?? ROUTES[0];

createRoot(document.getElementById('root')).render(
  <Layout rel={rel} activePath={route ? route.path : ''}>
    <page.component rel={rel} />
  </Layout>,
);
