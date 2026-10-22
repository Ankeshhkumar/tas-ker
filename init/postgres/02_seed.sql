-- Example data so the board is not empty on first run.

INSERT INTO projects (name, description) VALUES
  ('Website relaunch', 'New marketing site and docs'),
  ('Mobile app', 'First release of the companion app');

INSERT INTO tasks (title, description, project_id, status, priority, due_date) VALUES
  ('Write launch announcement', 'Blog post and email draft',        1, 'todo',        'medium', CURRENT_DATE + 7),
  ('Design new homepage',       'Hero, pricing and footer',         1, 'in_progress', 'high',   CURRENT_DATE + 3),
  ('Set up analytics',          'Add page-view tracking',           1, 'blocked',     'low',    NULL),
  ('Migrate blog posts',        'Move 40 posts to the new CMS',     1, 'done',        'medium', CURRENT_DATE - 2),
  ('Fix login crash on Android','Crash when the token expires',     2, 'in_progress', 'high',   CURRENT_DATE + 1),
  ('Add dark mode',             'Follow the system setting',        2, 'todo',        'low',    NULL),
  ('Publish to the app stores', 'Screenshots and store listing',    2, 'todo',        'high',   CURRENT_DATE + 14),
  ('Set up crash reporting',    'Capture and group crashes',        2, 'done',        'medium', CURRENT_DATE - 5);
