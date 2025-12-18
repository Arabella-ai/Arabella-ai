# Admin Panel Setup with Refine.dev

## Overview

Admin panel has been created using Refine.dev for managing templates and other resources.

## Structure

### Frontend
- **Location**: `/var/www/arabella/frontend/src/app/admin/`
- **Routes**:
  - `/admin` - Admin dashboard
  - `/admin/templates` - Template list
  - `/admin/templates/create` - Create new template
  - `/admin/templates/edit/[id]` - Edit template

### Backend
- **Admin Routes**: `/api/v1/admin/*`
- **Endpoints**:
  - `POST /api/v1/admin/templates` - Create template
  - `PUT /api/v1/admin/templates/:id` - Update template
  - `DELETE /api/v1/admin/templates/:id` - Delete template (soft delete)

## Features

1. **Template Management**:
   - List all templates
   - Create new templates
   - Edit existing templates
   - Delete templates (soft delete)

2. **Authentication**:
   - Admin routes require authentication
   - Uses existing JWT token system

## Access

1. Navigate to: `https://arabella.uz/admin`
2. Must be authenticated (login required)
3. Currently no role-based access control (all authenticated users can access)

## Future Enhancements

- Add role-based access control (admin vs regular user)
- Add user management
- Add video job management
- Add analytics dashboard
- Add settings management

## Deployment

The admin panel is included in the frontend build. After deploying:

```bash
cd /var/www/arabella/frontend
npm run build
sudo systemctl restart arabella-frontend
```

## Notes

- Refine.dev is used for the admin panel structure
- Custom data provider connects to your Go backend API
- Forms use react-hook-form with zod validation
- Styling matches your existing dark theme

