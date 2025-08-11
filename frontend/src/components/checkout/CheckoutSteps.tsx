import React from 'react';
import { CheckoutStep } from '@/app/checkout/page';

interface CheckoutStepsProps {
  currentStep: CheckoutStep;
}

const steps = [
  { id: 'shipping', name: 'Shipping', description: 'Delivery information' },
  { id: 'payment', name: 'Payment', description: 'Payment method' },
  { id: 'review', name: 'Review', description: 'Review & place order' },
];

export const CheckoutSteps: React.FC<CheckoutStepsProps> = ({ currentStep }) => {
  const getCurrentStepIndex = () => {
    return steps.findIndex(step => step.id === currentStep);
  };

  const currentStepIndex = getCurrentStepIndex();

  return (
    <div className="bg-white rounded-lg shadow p-6">
      <nav aria-label="Progress">
        <ol className="flex items-center">
          {steps.map((step, stepIdx) => {
            const isCompleted = stepIdx < currentStepIndex;
            const isCurrent = stepIdx === currentStepIndex;
            
            return (
              <li key={step.id} className={`relative ${stepIdx !== steps.length - 1 ? 'pr-8 sm:pr-20' : ''}`}>
                {/* Connector Line */}
                {stepIdx !== steps.length - 1 && (
                  <div className="absolute inset-0 flex items-center" aria-hidden="true">
                    <div className={`h-0.5 w-full ${isCompleted ? 'bg-blue-600' : 'bg-gray-200'}`} />
                  </div>
                )}
                
                {/* Step Circle */}
                <div className="relative flex items-center justify-center">
                  <div
                    className={`
                      flex h-10 w-10 items-center justify-center rounded-full border-2
                      ${isCompleted 
                        ? 'border-blue-600 bg-blue-600' 
                        : isCurrent 
                        ? 'border-blue-600 bg-white' 
                        : 'border-gray-300 bg-white'
                      }
                    `}
                  >
                    {isCompleted ? (
                      <svg className="h-6 w-6 text-white" fill="currentColor" viewBox="0 0 20 20">
                        <path fillRule="evenodd" d="M16.707 5.293a1 1 0 010 1.414l-8 8a1 1 0 01-1.414 0l-4-4a1 1 0 011.414-1.414L8 12.586l7.293-7.293a1 1 0 011.414 0z" clipRule="evenodd" />
                      </svg>
                    ) : (
                      <span className={`text-sm font-medium ${isCurrent ? 'text-blue-600' : 'text-gray-500'}`}>
                        {stepIdx + 1}
                      </span>
                    )}
                  </div>
                  
                  {/* Step Label */}
                  <div className="absolute top-12 left-1/2 transform -translate-x-1/2 text-center">
                    <div className={`text-sm font-medium ${isCurrent ? 'text-blue-600' : isCompleted ? 'text-gray-900' : 'text-gray-500'}`}>
                      {step.name}
                    </div>
                    <div className="text-xs text-gray-500 mt-1">
                      {step.description}
                    </div>
                  </div>
                </div>
              </li>
            );
          })}
        </ol>
      </nav>
    </div>
  );
};