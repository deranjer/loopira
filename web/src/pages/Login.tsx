import { Button, Paper, Stack, Text, TextInput, PasswordInput, Title } from '@mantine/core'
import { useForm } from '@mantine/form'
import { useNavigate } from 'react-router-dom'
import { ApiError } from '../lib/api/client'
import { useLogin } from '../lib/api/hooks'

export function Login() {
  const navigate = useNavigate()
  const login = useLogin()
  const form = useForm({ initialValues: { email: '', password: '' } })

  function submit(values: typeof form.values) {
    login.mutate(values, { onSuccess: () => navigate('/issues') })
  }

  const error =
    login.error instanceof ApiError ? login.error.message : login.isError ? 'Login failed' : null

  return (
    <div
      style={{
        minHeight: '100vh',
        display: 'flex',
        alignItems: 'center',
        justifyContent: 'center',
        background: 'var(--mantine-color-dark-7)',
      }}
    >
      <Paper w={420} p="xl" radius="md" bg="dark.6" style={{ border: '1px solid #1d1e21' }}>
        <form onSubmit={form.onSubmit(submit)}>
          <Stack gap="md">
            <div>
              <div style={{ width: 34, height: 34, borderRadius: 8, background: 'var(--mantine-color-accent-5)', marginBottom: 14 }} />
              <Title order={2} c="dark.0">
                Sign in to Loopira
              </Title>
            </div>
            <TextInput label="Email" placeholder="you@company.com" required {...form.getInputProps('email')} />
            <PasswordInput label="Password" required {...form.getInputProps('password')} />
            {error && (
              <Text c="red" size="sm">
                {error}
              </Text>
            )}
            <Button type="submit" fullWidth loading={login.isPending}>
              Sign in
            </Button>
          </Stack>
        </form>
      </Paper>
    </div>
  )
}
