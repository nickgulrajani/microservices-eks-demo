"""
User Service Microservice
Handles user authentication, registration, and profile management
"""

from fastapi import FastAPI, HTTPException, Depends, status
from fastapi.security import HTTPBearer, HTTPAuthorizationCredentials
from fastapi.middleware.cors import CORSMiddleware
from pydantic import BaseModel, EmailStr
from typing import Optional, List
import uvicorn
import os
import logging
from datetime import datetime, timedelta
import bcrypt
import jwt

# Configure logging
logging.basicConfig(level=logging.INFO)
logger = logging.getLogger(__name__)

app = FastAPI(
    title="User Service",
    description="Microservice for user management",
    version="1.0.0"
)

# CORS middleware
app.add_middleware(
    CORSMiddleware,
    allow_origins=["*"],  # In production, specify exact origins
    allow_credentials=True,
    allow_methods=["*"],
    allow_headers=["*"],
)

# Security
security = HTTPBearer()
SECRET_KEY = os.getenv("JWT_SECRET_KEY", "your-secret-key-change-in-production")
ALGORITHM = "HS256"

# Pydantic models
class UserCreate(BaseModel):
    email: EmailStr
    username: str
    password: str
    full_name: str

class UserResponse(BaseModel):
    id: int
    email: str
    username: str
    full_name: str
    created_at: datetime
    is_active: bool

class UserLogin(BaseModel):
    username: str
    password: str

class Token(BaseModel):
    access_token: str
    token_type: str

# Mock database (in production, use real database)
users_db = [
    {
        "id": 1,
        "email": "admin@example.com",
        "username": "admin",
        "full_name": "System Administrator",
        "hashed_password": bcrypt.hashpw("admin123".encode('utf-8'), bcrypt.gensalt()),
        "created_at": datetime.utcnow(),
        "is_active": True
    }
]

# Utility functions
def hash_password(password: str) -> bytes:
    return bcrypt.hashpw(password.encode('utf-8'), bcrypt.gensalt())

def verify_password(password: str, hashed_password: bytes) -> bool:
    return bcrypt.checkpw(password.encode('utf-8'), hashed_password)

def create_access_token(data: dict, expires_delta: Optional[timedelta] = None):
    to_encode = data.copy()
    if expires_delta:
        expire = datetime.utcnow() + expires_delta
    else:
        expire = datetime.utcnow() + timedelta(minutes=15)
    to_encode.update({"exp": expire})
    encoded_jwt = jwt.encode(to_encode, SECRET_KEY, algorithm=ALGORITHM)
    return encoded_jwt

def verify_token(credentials: HTTPAuthorizationCredentials = Depends(security)):
    try:
        payload = jwt.decode(credentials.credentials, SECRET_KEY, algorithms=[ALGORITHM])
        username: str = payload.get("sub")
        if username is None:
            raise HTTPException(
                status_code=status.HTTP_401_UNAUTHORIZED,
                detail="Could not validate credentials"
            )
        return username
    except jwt.PyJWTError:
        raise HTTPException(
            status_code=status.HTTP_401_UNAUTHORIZED,
            detail="Could not validate credentials"
        )

# API endpoints
@app.get("/health")
async def health_check():
    return {
        "status": "healthy",
        "service": "user-service",
        "timestamp": datetime.utcnow().isoformat(),
        "version": "1.0.0"
    }

@app.post("/register", response_model=UserResponse)
async def register_user(user: UserCreate):
    # Check if user already exists
    for existing_user in users_db:
        if existing_user["email"] == user.email or existing_user["username"] == user.username:
            raise HTTPException(
                status_code=status.HTTP_400_BAD_REQUEST,
                detail="User already exists"
            )
    
    # Create new user
    new_user = {
        "id": len(users_db) + 1,
        "email": user.email,
        "username": user.username,
        "full_name": user.full_name,
        "hashed_password": hash_password(user.password),
        "created_at": datetime.utcnow(),
        "is_active": True
    }
    
    users_db.append(new_user)
    
    # Return user without password
    return UserResponse(**{k: v for k, v in new_user.items() if k != "hashed_password"})

@app.post("/login", response_model=Token)
async def login(user_credentials: UserLogin):
    # Find user
    user = None
    for u in users_db:
        if u["username"] == user_credentials.username:
            user = u
            break
    
    if not user or not verify_password(user_credentials.password, user["hashed_password"]):
        raise HTTPException(
            status_code=status.HTTP_401_UNAUTHORIZED,
            detail="Incorrect username or password"
        )
    
    # Create access token
    access_token_expires = timedelta(minutes=30)
    access_token = create_access_token(
        data={"sub": user["username"]}, expires_delta=access_token_expires
    )
    
    return {"access_token": access_token, "token_type": "bearer"}

@app.get("/users/me", response_model=UserResponse)
async def get_current_user(current_username: str = Depends(verify_token)):
    user = None
    for u in users_db:
        if u["username"] == current_username:
            user = u
            break
    
    if not user:
        raise HTTPException(
            status_code=status.HTTP_404_NOT_FOUND,
            detail="User not found"
        )
    
    return UserResponse(**{k: v for k, v in user.items() if k != "hashed_password"})

@app.get("/users", response_model=List[UserResponse])
async def list_users(current_username: str = Depends(verify_token)):
    return [
        UserResponse(**{k: v for k, v in user.items() if k != "hashed_password"})
        for user in users_db
    ]

if __name__ == "__main__":
    uvicorn.run(
        "app:app",
        host="0.0.0.0",
        port=3001,
        reload=os.getenv("ENVIRONMENT") == "development"
    )
