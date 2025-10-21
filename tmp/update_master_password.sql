UPDATE users 
SET password = '$2a$12$AGqs4dCsv2hP4rZHz8mao.63kS3EOLTWLC7cmyfy8gR/3425Y.LXW' 
WHERE email = 'paulochiaradia72@gmail.com';

SELECT email, name, 
       CASE WHEN LENGTH(password) > 20 THEN 'Hash OK (length: ' || LENGTH(password) || ')' 
            ELSE 'Hash INVALID' 
       END as password_status
FROM users 
WHERE email = 'paulochiaradia72@gmail.com';
