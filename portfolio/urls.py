from django.urls import path
from . import views

urlpatterns = [
    path('', views.portfolio_view, name='portfolio'),
    path('api/contact/', views.contact_email, name='contact_email'),
]